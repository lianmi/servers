package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	gin_jwt_v2 "github.com/appleboy/gin-jwt/v2"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/common/helper"
	"github.com/lianmi/servers/internal/pkg/models"
	httpImpl "github.com/lianmi/servers/internal/pkg/transports/http"
	"go.uber.org/zap"
)

const (
	ErrorServerBusy = "server is busy"
	ErrorReLogin    = "relogin"
)

// Login form structure.
type Login struct {
	Username        string `form:"username" json:"username" binding:"required"`
	Password        string `form:"password" json:"password" binding:"required"`
	SmsCode         string `form:"smscode" json:"smscode" binding:"required"`
	DeviceID        string `form:"deviceid" json:"deviceid" binding:"required"`
	ClientType      int    `form:"clientype" json:"clientype" binding:"required"`
	Os              string `form:"os" json:"os" binding:"required"`
	ProtocolVersion string `form:"protocolversion" json:"protocolversion" binding:"required"`
	SdkVersion      string `form:"sdkversion" json:"sdkversion" binding:"required"`
	IsMaster        bool   `form:"ismaster" json:"ismaster"` //由于golang对false处理不对，所以不能设为必填
}

func ParseToken(tokenSrt string, SecretKey []byte) (claims jwt.Claims, err error) {
	var token *jwt.Token
	token, err = jwt.Parse(tokenSrt, func(*jwt.Token) (interface{}, error) {
		return SecretKey, nil
	})
	claims = token.Claims
	return
}

// lishijia 每增加一个路由需要在这里添加，并且在controllers/users.go及services/users.go增加相应的方法
func CreateInitControllersFn(
	pc *LianmiApisController,
) httpImpl.InitControllers {
	return func(r *gin.Engine) {
		r.POST("/register", pc.Register)              //注册用户
		r.POST("/resetpassword", pc.ResetPassword)    //重置密码， 可以不登录
		r.GET("/smscode/:mobile", pc.GenerateSmsCode) //根据手机生成短信注册码

		//TODO 增加JWT中间件
		authMiddleware, err := gin_jwt_v2.New(&gin_jwt_v2.GinJWTMiddleware{
			Realm:       "test zone",
			Key:         []byte(common.SecretKey),
			Timeout:     24 * 30 * time.Hour, //30日， common.ExpireTime, //expire过期时间   time.Hour
			MaxRefresh:  time.Hour,
			IdentityKey: common.IdentityKey,
			//构造JWT负载的回调
			PayloadFunc: func(data interface{}) gin_jwt_v2.MapClaims {
				//取出用户身份结构体里的数据
				if v, ok := data.(*models.UserRole); ok {
					//get roles from UserName
					v.UserRoles = pc.GetUserRoles(v.UserName)
					pc.logger.Debug("PayloadFunc",
						zap.String("UserName", v.UserName),
						zap.String("deviceID", v.DeviceID))

					jsonRole, _ := json.Marshal(v.UserRoles)
					//maps the claims in the JWT，将userRoles封装到JWT里
					return gin_jwt_v2.MapClaims{
						common.IdentityKey: v.UserName,           //用户账号
						"deviceID":         v.DeviceID,           //设备id
						"userRoles":        helper.B2S(jsonRole), //角色
					}
				} else {
					pc.logger.Error("Can not find Username")

				}
				return gin_jwt_v2.MapClaims{}
			},
			//解析并设置用户身份信息
			IdentityHandler: func(c *gin.Context) interface{} {

				roles := gin_jwt_v2.ExtractClaims(c)

				//extracts identity from roles
				jsonRole := roles["userRoles"].(string)

				var userRoles []*models.Role
				json.Unmarshal(helper.S2B(jsonRole), &userRoles)
				//Set the identity
				log.Println("IdentityHandler run ... %#v", roles)

				return &models.UserRole{
					UserName:  roles[common.IdentityKey].(string),
					DeviceID:  roles["deviceID"].(string),
					UserRoles: userRoles,
				}
			},
			//认证，根据登录信息对用户进行身份验证的回调函数
			Authenticator: func(c *gin.Context) (interface{}, error) {
				pc.logger.Debug("Authenticator ...")
				//handles the login logic. On success LoginResponse is called, on failure Unauthorized is called
				var loginVals Login //重要！不能用models.User，因为有很多必填字段
				if err := c.ShouldBind(&loginVals); err != nil {
					pc.logger.Error("Authenticator Error", zap.Error(err))
					return "", gin_jwt_v2.ErrMissingLoginValues
				}
				isMaster := loginVals.IsMaster
				smscode := loginVals.SmsCode
				username := loginVals.Username
				password := loginVals.Password
				deviceID := loginVals.DeviceID
				clientType := loginVals.ClientType
				os := loginVals.Os

				// 检测用户是否可以登陆
				if pc.CheckUser(isMaster, smscode, username, password, deviceID, os, clientType) {
					pc.logger.Debug("Authenticator , CheckUser .... true")
					return &models.UserRole{
						UserName: username,
						DeviceID: deviceID,
					}, nil

				}

				return nil, gin_jwt_v2.ErrFailedAuthentication
			},

			//授权, 接收用户信息并编写授权规则，本项目的API权限控制就是通过该函数编写授权规则的
			Authorizator: func(data interface{}, c *gin.Context) bool {
				pc.logger.Debug("Authorizator ...授权")

				token := gin_jwt_v2.GetToken(c)

				claims, err := ParseToken(token, []byte(common.SecretKey))
				if nil != err {
					pc.logger.Error("ParseToken Error", zap.Error(err))
				}
				userName := claims.(jwt.MapClaims)[common.IdentityKey].(string)
				deviceID := claims.(jwt.MapClaims)["deviceID"].(string)

				pc.logger.Debug("Authorizator", zap.String("userName", userName), zap.String("deviceID", deviceID))

				//检测deviceID 对应的此令牌是否存在redis里，如果不存在，则不能授权通过
				isExists := pc.ExistsTokenInRedis(deviceID, token)
				if !isExists {
					pc.logger.Debug("此deviceID的令牌不存在redis里", zap.String("deviceID", deviceID))
					return false
				}

				//暂时只有jwt有效，都放行
				return true

				// if v, ok := data.(*models.UserRole); ok {
				// 	for _, itemRole := range v.UserRoles {

				// 		pc.logger.Debug("jwt携带的用户角色", zap.String("itemRole.Value", itemRole.Value))

				// 		if itemRole.Value == "admin" { //超级管理员，目前只支持一种后台管理用户
				// 			return true
				// 		}
				// 	}
				// }
				// pc.logger.Debug("Authorizator faild, must be admin")

				// return false
			},
			//处理不进行授权的逻辑
			Unauthorized: func(c *gin.Context, code int, message string) {
				c.JSON(code, gin.H{
					"code":    code,
					"message": message,
				})
			},
			LoginResponse: func(c *gin.Context, code int, token string, t time.Time) {
				//解析JWT令牌
				claims, err := ParseToken(token, []byte(common.SecretKey))
				if nil != err {
					pc.logger.Error("ParseToken Error", zap.Error(err))
				}
				userName := claims.(jwt.MapClaims)[common.IdentityKey].(string)
				deviceID := claims.(jwt.MapClaims)["deviceID"].(string)
				pc.logger.Debug("LoginResponse", zap.String("userName", userName), zap.String("deviceID", deviceID), zap.String("expire", t.Format(time.RFC3339)))

				//将userName, deviceID, token及expire保存到redis, 用于mqtt协议的消息的授权验证
				pc.SaveUserToken(userName, deviceID, token, t)

				//向客户端回复生成的JWT令牌
				RespData(c, http.StatusOK, code, token)

			},
			// TokenLookup is a string in the form of "<source>:<name>" that is used
			// to extract token from the request.
			// Optional. Default value "header:Authorization".
			// Possible values:
			// - "header:<name>"
			// - "query:<name>"
			// - "cookie:<name>"
			// - "param:<name>"
			TokenLookup: "header: Authorization, query: token, cookie: jwt",
			// TokenLookup: "query:token",
			// TokenLookup: "cookie:token",

			// TokenHeadName is a string in the header. Default value is "Bearer"
			TokenHeadName: "Bearer",

			// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
			TimeFunc: time.Now,
		})

		if err != nil {
			log.Fatal("JWT Error:" + err.Error())
		}

		r.POST("/login", authMiddleware.LoginHandler)

		r.NoRoute(authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
			// claims := gin_jwt_v2.ExtractClaims(c)
			// log.Printf("NoRoute claims: %#v\n", claims)
			c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
		})

		//=======无须登录也能访问的uri==========/
		authNone := r.Group("/shops")
		authNone.GET("/nearby", pc.QueryShopsNearby)

		//=======用户模块==========/
		authUser := r.Group("/v1/user") //带v1的路由都必须使用Bearer JWT 才能正常访问-普通用户及后台操作人员都能访问
		// Refresh time can be longer than token timeout
		authUser.GET("/refresh_token", authMiddleware.RefreshHandler)
		authUser.Use(authMiddleware.MiddlewareFunc())
		{
			authUser.GET("/signout", pc.SignOut)            //登出
			authUser.GET("/info/:id", pc.GetUser)           //根据id获取用户信息
			authUser.POST("/chanpassword", pc.ChanPassword) //修改(重置)用户密码

		}

		//=======好友模块==========/
		authFriend := r.Group("/v1/friend")
		authFriend.Use(authMiddleware.MiddlewareFunc())
		{

		}

		//=======群组模块==========/
		authTeam := r.Group("/v1/team")
		authTeam.Use(authMiddleware.MiddlewareFunc())
		{

			//4-2 获取群组成员信息
			// authTeam.GET("/teammembers/:teamid", pc.GetTeamMembers)

			// 4-3 查询群信息
			// authTeam.GET("/getteam/:teamid", pc.GetTeam)

			//4-24 获取指定群组成员
			// authTeam.POST("/pullteammembers", pc.PullTeamMembers)

			// 4-27 分页获取群成员信息
			// authTeam.POST("/getteammemberspage", pc.GetTeamMembersPage)

			//5-1发送吸顶式群消息
			// authTeam.POST("/sendteamroofmsg", pc.SendTeamRoofMsg)
		}

		//=======商品模块==========/
		authProduct := r.Group("/v1/product")
		authProduct.Use(authMiddleware.MiddlewareFunc())
		{
			//查询通用商品by productid
			authProduct.GET("/getgeneralproduct/:productid", pc.GetGeneralProductByID)

			//查询通用商品分页-按商品种类查询, /getgeneralproductspage?producttype=1
			authProduct.GET("/getgeneralproductspage", pc.GetGeneralProductPage)

		}

		//=======订单模块==========/
		authOrder := r.Group("/v1/order")
		authOrder.Use(authMiddleware.MiddlewareFunc())
		{
		}

		//=======钱包模块==========/
		authWallet := r.Group("/v1/wallet")
		authWallet.Use(authMiddleware.MiddlewareFunc())
		{
		}

		//=======会员付费分销模块==========/
		authMembership := r.Group("/v1/membership")
		authMembership.Use(authMiddleware.MiddlewareFunc())
		{
			//查询商户会员营销方式
			authMembership.GET("/getmembershipcardsalemode", pc.GetMembershipCardSaleMode)

			// 设置商户会员营销方式
			authMembership.POST("/Setmembershipcardsalemode", pc.SetMembershipCardSaleMode)

			//为自己或他人支付会员费
			authMembership.POST("/payformembership", pc.PayForMembership)

			//商户查询当前名下用户总数，按月统计付费会员总数及返佣金额，是否已经返佣
			authMembership.GET("/getbusinessmembership", pc.GetBusinessMembership)
		}

		//=======客服模块==========/
		authCustomerService := r.Group("/v1/customerservice")
		authCustomerService.Use(authMiddleware.MiddlewareFunc())
		{
			//查询在线客服id数组
			authCustomerService.GET("/querycustomerservices", pc.QueryCustomerServices)

			//查询评分
			authCustomerService.GET("/querygrades", pc.QueryGrades)

			//客服增加评分标题及内容
			authCustomerService.GET("/addgrade", pc.AddGrade)

			//用户提交评分
			authCustomerService.GET("/submitgrade", pc.SubmitGrade)
		}

		//=======后台各个功能模块==========/
		adminAuth := r.Group("/admin") //带/admin的路由都必须使用Bearer JWT，并且Role为admin才能正常访问
		adminAuth.Use(authMiddleware.MiddlewareFunc())
		{

			//根据用户账号, 将此用户封号
			adminAuth.GET("/blockuser/:username", pc.BlockUser)

			//根据用户账号，将此用户解封
			adminAuth.GET("/disblockuser/:username", pc.DisBlockUser)

			//授权新创建的群组
			adminAuth.GET("/approveteam/:teamid", pc.ApproveTeam)

			//封禁群组
			adminAuth.GET("/blockteam/:teamid", pc.BlockTeam)

			//解封群组
			adminAuth.GET("/disblockteam/:teamid", pc.DisBlockTeam)

			//增加通用商品
			adminAuth.POST("/addgeneralproduct", pc.AddGeneralProduct)

			//修改通用商品
			adminAuth.POST("/updategeneralproduct", pc.UpdateGeneralProduct)

			//删除通用商品 by productid
			adminAuth.DELETE("/generalproduct/:productid", pc.DeleteGeneralProduct)

			//增加在线客服id
			adminAuth.GET("/addcustomerservice", pc.AddCustomerService)

			//删除 在线客服id
			adminAuth.GET("/deletecustomerservice", pc.DeleteCustomerService)

			//编辑 在线客服id
			adminAuth.GET("/updatecustomerservice", pc.UpdateCustomerService)

		}
	}
}

var ProviderSet = wire.NewSet(NewLianmiApisController, CreateInitControllersFn)

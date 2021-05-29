package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/iGoogle-ink/gopay/wechat/v3"

	gin_jwt_v2 "github.com/appleboy/gin-jwt/v2"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/common"
	LMCError "github.com/lianmi/servers/internal/pkg/lmcerror"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/lianmi/servers/internal/common/helper"
	"github.com/lianmi/servers/internal/pkg/models"
	httpImpl "github.com/lianmi/servers/internal/pkg/transports/http"
	"github.com/lianmi/servers/util/conv"
	"go.uber.org/zap"
)

const (
	ErrorServerBusy = "server is busy"
	ErrorReLogin    = "relogin"
)

// Login form structure.
type Login struct {
	Mobile          string `form:"mobile" json:"mobile"`                        // 11位手机号 选填
	Username        string `form:"username" json:"username"`                    //注册号，当mobile非空的时候，选填
	Password        string `form:"password" json:"password" `                   //密码 Username非空时必填
	SmsCode         string `form:"smscode" json:"smscode" `                     //短信校验码，Mobile非空时必填
	DeviceID        string `form:"deviceid" json:"deviceid" binding:"required"` //必填
	UserType        int    `form:"usertype" json:"usertype" binding:"required"` //必填，用户类型，区分普通用户或商户
	Os              string `form:"os" json:"os" `                               //非必填，客户端的操作系统
	ProtocolVersion string `form:"protocolversion" json:"protocolversion"`      //非必填，协议版本
	SdkVersion      string `form:"sdkversion" json:"sdkversion"`                //非必填，sdk版本
	IsMaster        bool   `form:"ismaster" json:"ismaster"`                    //由于golang对false处理不对，所以不能设为必填
}

type LoginResp struct {
	Username    string `form:"username" json:"username"`         // 注册账号
	UserType    int    `form:"user_type" json:"user_type"`       // 用户类型 1-普通，2-商户
	State       int    `form:"state" json:"state"`               // 普通用户： 0-非VIP 1-付费用户(购买会员) 2-封号  商户：0-预审核, 1-审核通过, 2 -占位, 3-审核中
	AuditResult string `form:"audit_result" json:"audit_result"` // 商户：  当state=3，此字段是审核的文字报告，如审核中，地址不符，照片模糊等
	JwtToken    string `form:"jwt_token" json:"jwt_token"`       // 令牌
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

	var errPay error
	pc.payWechat, errPay = wechat.NewClientV3(
		common.WechatPay_appID,
		common.WechatPay_mchId,
		common.WechatPay_serierNo,
		common.WechatPay_apiV3Key,
		common.Wechat_pkContent)

	if errPay != nil {
		pc.logger.Warn("微信支付初始化失败")
		// 失败不影响其他业务
	} else {
		pc.logger.Debug("微信支付初始化成功")
	}

	//if err != nil {
	//	fmt.Println("微信支付客户端初始化失败")
	//	return
	//}

	return func(r *gin.Engine) {
		//以下路由不做鉴权
		r.POST("/register", pc.Register)                              //注册用户
		r.POST("/resetpassword", pc.ResetPassword)                    //重置密码， 可以不登录， 但必须用短信校验码
		r.GET("/smscode/:mobile", pc.GenerateSmsCode)                 //根据手机生成短信注册码
		r.POST("/validatecode", pc.ValidateCode)                      //验证码验证接口
		r.GET("/getusernamebymobile/:mobile", pc.GetUsernameByMobile) //根据手机号获取注册账号id

		authMiddleware, err := gin_jwt_v2.New(&gin_jwt_v2.GinJWTMiddleware{
			Realm:       "gin",
			Key:         []byte(common.SecretKey),
			Timeout:     24 * 30 * time.Hour, //30日， common.ExpireTime, //expire过期时间   time.Hour
			MaxRefresh:  time.Hour,
			IdentityKey: common.IdentityKey,
			//构造JWT负载的回调
			PayloadFunc: func(data interface{}) gin_jwt_v2.MapClaims {
				//取出用户身份结构体里的数据
				if v, ok := data.(*models.UserRole); ok {
					// 根据UserName获取用户角色
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
				//log.Println("IdentityHandler run ... %#v", roles)

				return &models.UserRole{
					UserName:  roles[common.IdentityKey].(string),
					DeviceID:  roles["deviceID"].(string),
					UserRoles: userRoles,
				}
			},
			//认证，根据登录信息对用户进行身份验证的回调函数
			Authenticator: func(c *gin.Context) (interface{}, error) {
				//handles the login logic. On success LoginResponse is called, on failure Unauthorized is called
				var loginVals Login //重要！不能用models.User，因为有很多必填字段
				if err := c.ShouldBind(&loginVals); err != nil {
					pc.logger.Error("Authenticator Error", zap.Error(err))
					return "", gin_jwt_v2.ErrMissingLoginValues
				}
				var err error
				// isMaster := loginVals.IsMaster
				isMaster := true //强制设置为主设备
				smscode := strings.TrimSpace(loginVals.SmsCode)
				mobile := strings.TrimSpace(loginVals.Mobile)
				username := strings.TrimSpace(loginVals.Username)
				password := strings.TrimSpace(loginVals.Password)
				deviceID := strings.TrimSpace(loginVals.DeviceID)
				userType := loginVals.UserType //1-普通用户 2-商户
				os := strings.TrimSpace(loginVals.Os)

				pc.logger.Debug("Authenticator ...",
					zap.String("mobile", mobile),
					zap.String("username", username),
					zap.String("password", password),
					zap.String("smscode", smscode),
					zap.String("deviceID", deviceID),
					zap.Int("userType", userType),
					zap.String("os", os),
				)

				if mobile == "" && username == "" {
					pc.logger.Error("Mobile and Username are both empty")
					return "", gin_jwt_v2.ErrMissingLoginValues
				} else if mobile != "" && smscode == "" {
					pc.logger.Error("SmsCode is empty")
					return "", gin_jwt_v2.ErrMissingLoginValues
				} else if mobile != "" && username == "" {
					//不是手机
					if len(mobile) != 11 {
						pc.logger.Warn("Mobile error", zap.Int("len", len(mobile)))

						return "", gin_jwt_v2.ErrMissingLoginValues
					}

					//不是全数字
					if !conv.IsDigit(mobile) {
						pc.logger.Warn("Mobile Is not Digit")
						return "", gin_jwt_v2.ErrMissingLoginValues
					}
					//检测校验码是否正确
					if !pc.service.CheckSmsCode(mobile, smscode) {
						pc.logger.Error("CheckSmsCode error, SmsCode is wrong")

						errMsg := LMCError.ErrorMsg(LMCError.SmsCodeCheckError)
						return "", errors.New(errMsg)
					}

					//根据手机号获取用户id ???
					username, err = pc.service.GetUsernameByMobile(mobile)
					if err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							pc.logger.Warn("mobile is not registered")

							//将用户注册
							// return "", gin_jwt_v2.ErrMissingLoginValues
							user := models.User{
								UserBase: models.UserBase{
									Mobile:    mobile,   //注册手机
									AllowType: 3,        //用户加好友枚举，默认是3
									UserType:  userType, //用户类型 1-普通，2-商户
									State:     0,        //状态 0-普通用户，非VIP 1-付费用户(购买会员) 2-封号
								},
							}

							if username, err = pc.service.Register(&user); err == nil {
								pc.logger.Debug("Register user success", zap.String("username", username))
								// 检测用户是否可以登录, true-可以允许登录
								if pc.CheckUser(true, username, password, deviceID, os, userType) {
									pc.logger.Debug("Authenticator , CheckUser .... true")

									return &models.UserRole{
										UserName: username,
										DeviceID: deviceID,
									}, nil

								} else {
									pc.logger.Warn("Authenticator , CheckUser .... false")
									return "", gin_jwt_v2.ErrMissingLoginValues
								}

							} else {
								pc.logger.Error("Register user error", zap.Error(err))
								pc.logger.Warn("Authenticator , CheckUser .... false")
							}
						}
						pc.logger.Warn("GetUsernameByMobile error")
						errMsg := LMCError.ErrorMsg(LMCError.MobileNotRegisterError)
						return "", errors.Wrap(err, errMsg)
					}

					//如果最终username为空则未注册
					if username != "" {

						//检测校验码是否正确
						if pc.LoginBySmscode(username, mobile, smscode, deviceID, os, userType) {
							pc.logger.Debug("Authenticator , LoginBySmsCode .... true")

							return &models.UserRole{
								UserName: username,
								DeviceID: deviceID,
							}, nil
						} else {
							pc.logger.Warn("Authenticator , LoginBySmsCode .... false")
						}
					}

				} else if mobile == "" && username != "" {
					if password == "" {
						pc.logger.Error("password is empty")
						return "", gin_jwt_v2.ErrMissingLoginValues
					}
					mobile, err = pc.service.GetMobileByUsername(username)
					if err != nil {
						pc.logger.Warn("GetMobileByUsername error")
						return "", gin_jwt_v2.ErrMissingLoginValues
					}

					if mobile == "" {
						pc.logger.Warn("mobile get error")
						return "", gin_jwt_v2.ErrMissingLoginValues
					}
					// 检测用户是否可以登录, true-可以允许登录
					if pc.CheckUser(isMaster, username, password, deviceID, os, userType) {
						pc.logger.Debug("Authenticator , CheckUser .... true")

						return &models.UserRole{
							UserName: username,
							DeviceID: deviceID,
						}, nil

					} else {
						pc.logger.Warn("Authenticator , CheckUser .... false")
					}

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
					return false
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

				//只有jwt有效，都放行
				return true
			},

			//处理不进行授权的逻辑
			Unauthorized: func(c *gin.Context, code int, message string) {
				pc.logger.Debug("Unauthorized",
					zap.Int("code", code),
					zap.String("message", message),
				)

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
					return
				}
				userName := claims.(jwt.MapClaims)[common.IdentityKey].(string)
				deviceID := claims.(jwt.MapClaims)["deviceID"].(string)

				//将userName, deviceID, token及expire保存到redis, 用于mqtt协议的消息的授权验证
				pc.SaveUserToken(userName, deviceID, token, t)

				//将新的token写入redis（db=6）
				pc.SetMqttBrokerRedisAuth(deviceID, token)

				user, err := pc.service.GetUser(userName)
				if err != nil {
					pc.logger.Error("Get User by userName error", zap.Error(err))
					RespData(c, http.StatusOK, 500, "Get User by userName error")
					return
				}

				var auditResult string
				if int(user.User.UserType) == 2 && int(user.User.State) == 3 {
					auditResult = "商户进驻已受理, 审核中..."
				}
				pc.logger.Debug("LoginResponse",
					zap.String("userName", userName),
					zap.String("deviceID", deviceID),
					zap.Int("UserType", int(user.User.UserType)),
					zap.Int("State", int(user.User.State)),
					zap.String("expire", t.Format(time.RFC3339)),
				)

				//向客户端回复注册号及生成的JWT令牌，  用户类型，用户状态，审核结果
				RespData(c, http.StatusOK, code, &LoginResp{
					Username:    userName,
					UserType:    int(user.User.UserType),
					State:       int(user.User.State),
					AuditResult: auditResult,
					JwtToken:    token,
				})

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

		// github的OAuth授权登录回调uri
		r.GET("/login-github", pc.GitHubOAuth)

		// 获取App下载的链接url
		r.GET("/getdownloadurl", pc.GetDownloadURL)

		r.POST("/login", authMiddleware.LoginHandler)

		r.NoRoute(authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
			// claims := gin_jwt_v2.ExtractClaims(c)
			// log.Printf("NoRoute claims: %#v\n", claims)
			c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
		})

		//=======鉴权授权模块==========/
		auth := r.Group("/v1") //带v1的路由都必须使用Bearer JWT 才能正常访问-普通用户及后台操作人员都能访问
		// Refresh time can be longer than token timeout
		auth.GET("/refresh_token", authMiddleware.RefreshHandler)
		auth.Use(authMiddleware.MiddlewareFunc())
		{

			auth.GET("/devices", pc.GetAllDevices) //获取当前用户的登录设备
			auth.GET("/signout", pc.SignOut)       //登出

		}

		//=======用户模块==========/
		userGroup := r.Group("/v1/user") //带v1的路由都必须使用Bearer JWT 才能正常访问-普通用户及后台操作人员都能访问
		userGroup.Use(authMiddleware.MiddlewareFunc())
		{
			userGroup.GET("/getuser/:id", pc.GetUser)  //根据用户注册号获取用户详细 信息,  如果是本身，则返回更加详尽的信息，包括到期时间
			userGroup.POST("/list", pc.QueryUsers)     //多条件不定参数批量分页获取用户列表
			userGroup.GET("/likes", pc.UserLikes)      //获取当前用户对所有店铺点赞情况
			userGroup.POST("/getuserdb", pc.GetUserDb) //根据用户注册号获取用户数据库

		}

		//=======店铺模块==========/
		storeGroup := r.Group("/v1/store") //带v1的路由都必须使用Bearer JWT 才能正常访问-普通用户及后台操作人员都能访问
		storeGroup.Use(authMiddleware.MiddlewareFunc())
		{
			storeGroup.GET("/storeinfo/:id", pc.GetStore)        //根据商户注册id获取店铺资料
			storeGroup.GET("/types", pc.GetStoreTypes)           //返回商品种类
			storeGroup.POST("/savestore", pc.AddStore)           //增加或修改店铺资料
			storeGroup.POST("/list", pc.QueryStoresNearby)       // 不定条件查询店铺列表
			storeGroup.POST("/productslist", pc.GetProductsList) //获取某个商户的所有商品列表
			storeGroup.POST("/productlists", pc.GetStoreProductLists)
			storeGroup.GET("/likes/:id", pc.StoreLikes)           //获取店铺的所有点赞用户列表
			storeGroup.GET("/likescount/:id", pc.StoreLikesCount) //获取店铺的点赞总数
			storeGroup.POST("/like/:id", pc.ClickLike)            //对某个店铺进行点赞
			storeGroup.DELETE("/like/:id", pc.DeleteClickLike)    //取消对某个店铺点赞
			storeGroup.GET("/islike/:id", pc.GetIsLike)           //判断当前用户是否对某个店铺点过赞
			storeGroup.POST("/defaultopk", pc.DefaultOPK)         //设置当前商户的默认OPK

		}

		//=======好友模块==========/
		friendGroup := r.Group("/v1/friend")
		friendGroup.Use(authMiddleware.MiddlewareFunc())
		{

		}

		//=======群组模块==========/
		teamGroup := r.Group("/v1/team")
		teamGroup.Use(authMiddleware.MiddlewareFunc())
		{

		}

		//=======商品模块==========/
		productGroup := r.Group("/v1/product")
		productGroup.Use(authMiddleware.MiddlewareFunc())
		{

			//查询通用商品的详情 by productid
			productGroup.GET("/generalproduct/:productid", pc.GetGeneralProductByID)
			//根据商户注册号查询所有上架商品
			productGroup.POST("/productslist", pc.GetProductsList)
			//根据商品ID获取商品详情
			productGroup.GET("/info/:productid", pc.GetProductInfo)
			//设置商品的子类型
			productGroup.POST("/setsubtype", pc.SetProductSubType)
			// 获取通用商品id列表
			productGroup.GET("/generalproducts", pc.GetGeneralProjectIDs)
			productGroup.GET("/generalproductslist", pc.GetGeneralProjectsList)
		}

		//=======订单模块==========/
		orderGroup := r.Group("/v1/order")
		orderPubGroup := r.Group("/v1/order")
		orderGroup.Use(authMiddleware.MiddlewareFunc())
		{
			//商户端: 将完成订单拍照所有图片上链
			orderGroup.POST("/uploadorderimage", pc.UploadOrderImages)

			//商户端: 将订单body经过RSA加密后提交到服务端
			orderGroup.POST("/uploadorderbody", pc.UploadOrderBody)

			//用户端: 根据 OrderID 获取所有订单拍照图片
			orderGroup.GET("/orderimage/:orderid", pc.DownloadOrderImage)

			// 用户向商户发起订单
			orderGroup.POST("/pay", pc.OrderPayToBusiness)
			orderGroup.POST("/get_order_rate", pc.OrderCalcPrice)
			// 翻页获取订单列表
			orderGroup.GET("/lists", pc.OrderGetLists)
			// 通过订单id 获取订单信息接口
			orderGroup.GET("/info/:id", pc.OrderGetOrderInfoByID)
			// 通过订单id 修改订单状态
			orderGroup.POST("/update_status", pc.OrderUpdateStatusByOrderID)
			// 微信支付回调接口
			r.POST("/callback/wechat/notify", pc.OrderWechatCallbackRelease)
			// TODO 模拟设置微信支付回调的接口
			orderPubGroup.POST("/callback/wechat_test", pc.OrderWechatCallback)
			// 推送兑奖金额
			orderGroup.POST("/push_prize", pc.OrderPushPrize)
			// 通过订单id 查找微信的交易信息
			orderGroup.GET("/wechat_transactions/:orderid", pc.OrderFindWechatTransactions)
			//当前用户删除自己的订单
			orderGroup.POST("/delete/:id", pc.OrderDeleteByUserIDAndOrderID)
			// 删除当前用户的所有订单
			orderGroup.POST("/clearall", pc.OrderDeleteByUserID)
			// 通过关键字搜索订单
			orderGroup.POST("/search", pc.OrderSerachByKeyWord)

			//
			//orderPubGroup.POST("/wechat/callback", pc.OrderWechatCallback)
		}

		//=======钱包模块==========/
		// walletGroup := r.Group("/v1/wallet")
		// walletGroup.Use(authMiddleware.MiddlewareFunc())
		// {
		// 	//支付宝预支付动作
		// 	walletGroup.POST("/alipay", pc.PreAlipay)
		// 	//支付宝回调
		// 	walletGroup.POST("/alipay/callback", pc.AlipayCallback)
		// 	//支付宝回调
		// 	walletGroup.POST("/alipay/notify", pc.AlipayNotify)

		// 	//微信预支付动作
		// 	walletGroup.POST("/wxpay", pc.PreWXpay)

		// 	//微信回调
		// 	walletGroup.POST("/wxpaynotify", pc.WXpayNotify)

		// }

		//=======会员付费分销模块==========/
		membershipGroup := r.Group("/v1/membership")
		membershipGroup.Use(authMiddleware.MiddlewareFunc())
		{

			//查询VIP会员价格表
			membershipGroup.GET("/pricelist", pc.GetVipPriceList)

			//商户查询当前名下用户总数，按月统计付费会员总数及返佣金额，是否已经返佣
			membershipGroup.GET("/getall", pc.GetBusinessMembership)

			//统计用户佣金统计
			membershipGroup.PUT("/statistics", pc.CommissonSatistics)

			//用户查询按月统计发展的付费会员总数及返佣金额，是否已经返佣
			membershipGroup.GET("/commssions", pc.GetCommissionStatistics)

			//提交佣金提现申请
			membershipGroup.POST("/submitcommssionwithdraw", pc.SubmitCommssionWithdraw)

		}

		//=======客服模块==========/
		customerServiceGroup := r.Group("/v1/customerservice")
		customerServiceGroup.Use(authMiddleware.MiddlewareFunc())
		{
			//查询在线客服id数组
			customerServiceGroup.GET("/querycustomerservices", pc.QueryCustomerServices)

			//查询评分
			customerServiceGroup.GET("/querygrades", pc.QueryGrades)

			//客服增加评分标题及内容
			customerServiceGroup.POST("/addgrade", pc.AddGrade)

			//用户提交评分
			customerServiceGroup.POST("/submitgrade", pc.SubmitGrade)
		}

		//=======后台各个功能模块==========/
		adminGroup := r.Group("/admin") //带/admin的路由都必须使用Bearer JWT，并且Role为admin才能正常访问
		adminGroup.Use(authMiddleware.MiddlewareFunc())
		{

			//根据用户账号, 将此用户封号
			adminGroup.POST("/blockuser/:username", pc.BlockUser)

			//根据用户账号，将此用户解封
			adminGroup.POST("/disblockuser/:username", pc.DisBlockUser)

			//授权新创建的群组
			adminGroup.GET("/approveteam/:teamid", pc.ApproveTeam)

			//封禁群组
			adminGroup.POST("/blockteam/:teamid", pc.BlockTeam)

			//解封群组
			adminGroup.POST("/disblockteam/:teamid", pc.DisBlockTeam)

			//查询通用商品分页-按商品种类查询
			adminGroup.POST("/generalproductslist", pc.GetGeneralProductPage)

			//增加通用商品
			adminGroup.POST("/addgeneralproduct", pc.AddGeneralProduct)

			//修改通用商品
			adminGroup.POST("/updategeneralproduct", pc.UpdateGeneralProduct)

			//删除通用商品 by productid
			adminGroup.DELETE("/generalproduct/:productid", pc.DeleteGeneralProduct)

			//增加在线客服id
			adminGroup.POST("/addcustomerservice", pc.AddCustomerService)

			//删除在线客服id
			adminGroup.DELETE("/deletecustomerservice/:id", pc.DeleteCustomerService)

			//修改在线客服资料
			adminGroup.POST("/updatecustomerservice", pc.UpdateCustomerService)

			//将店铺审核通过
			adminGroup.POST("/auditstore", pc.AuditStore)

			// 增加商户支持的彩种
			adminGroup.POST("/addstoreproduct", pc.AdminAddStoreProductItem)
			// 查看现有缓存的key
			adminGroup.GET("/allcachekey", pc.AdminFindAllCacheLKey)

			// 查看缓存的key 的内容
			adminGroup.GET("/cachevalue/:key", pc.AdminGetCacheKeyValue)

			// 删除指定的缓存key
			adminGroup.DELETE("/cachevalue/:key", pc.AdminDelCacheKeyValue)

			//导入广东省彩票网点excel文件
			adminGroup.POST("/loadexcel", pc.LoadExcel)

			//查询广东省彩票网点记录
			adminGroup.GET("/lotterystores", pc.GetLotteryStores)

			adminGroup.POST("/addstore", pc.AdminAddStore)

			//批量增加网点
			adminGroup.POST("/batchaddstores", pc.BatchAddStores)

			//批量设置网点opk
			adminGroup.POST("/admindefaultopk", pc.AdminDefaultOPK)

		}
	}
}

var ProviderSet = wire.NewSet(NewLianmiApisController, CreateInitControllersFn)

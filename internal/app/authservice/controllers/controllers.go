package controllers

import (
	// "fmt"
	// "errors"
	"net/http"
	"encoding/json"
	"time"
	"log"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"go.uber.org/zap"
	"github.com/lianmi/servers/internal/pkg/models"
	httpImpl "github.com/lianmi/servers/internal/pkg/transports/http"
	gin_jwt_v2 "github.com/appleboy/gin-jwt/v2"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/lianmi/servers/internal/common/helper"
)
const (
    ErrorServerBusy = "server is busy"
    ErrorReLogin = "relogin"
)

var (
    SecretKey = "lianimicloud-secret"  //salt
    ExpireTime = 3600  //token expire time
)
// Login form structure.
type Login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

var identityKey = "userName"

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
	pc *UsersController,
) httpImpl.InitControllers {
	return func(r *gin.Engine) {
		r.POST("/register", pc.Register)      //注册用户
		r.GET("/smscode", pc.GenerateSmsCode) //根据手机生成短信注册码

		//TODO 增加JWT中间件
		authMiddleware, err := gin_jwt_v2.New(&gin_jwt_v2.GinJWTMiddleware{
			Realm:       "test zone",
			Key:         []byte(SecretKey),
			Timeout:     time.Hour,
			MaxRefresh:  time.Hour,
			IdentityKey: identityKey,
			//登录期间的回调的函数
			PayloadFunc: func(data interface{}) gin_jwt_v2.MapClaims {
				//TODO 这里仅仅判断了User是否存在就授权，要改进
				if v, ok := data.(*models.UserRole); ok {
					//get roles from UserName
					v.UserRoles = pc.GetUserRoles(v.UserName) 
					pc.logger.Debug("Find Username", zap.String("UserName", v.UserName))

					jsonRole, _ := json.Marshal(v.UserRoles)
					//maps the claims in the JWT
					return gin_jwt_v2.MapClaims{
						identityKey:  v.UserName,
						"userRoles": helper.B2S(jsonRole),
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
				// log.Println("IdentityHandler run ... %#v", roles)
				
				return &models.UserRole{
					UserName:  roles[identityKey].(string),
					UserRoles: userRoles,
				}
			},
			//根据登录信息对用户进行身份验证的回调函数
			Authenticator: func(c *gin.Context) (interface{}, error) {
				pc.logger.Debug("Authenticator ...")
				//handles the login logic. On success LoginResponse is called, on failure Unauthorized is called
				var loginVals Login //重要！不能用models.User，因为有很多必填字段
				if err := c.ShouldBind(&loginVals); err != nil {
					pc.logger.Error("Authenticator Error", zap.Error(err))
					return "", gin_jwt_v2.ErrMissingLoginValues
				}
				username := loginVals.Username
				password := loginVals.Password

				// 检测用户及密码是否存在
				if pc.CheckUser(username, password) {
					pc.logger.Debug("Authenticator , CheckUser .... true")
					return &models.UserRole {
						UserName: username,
					}, nil
				}

				return nil, gin_jwt_v2.ErrFailedAuthentication
			},

			//接收用户信息并编写授权规则，本项目的API权限控制就是通过该函数编写授权规则的
			Authorizator: func(data interface{}, c *gin.Context) bool {
				if v, ok := data.(*models.UserRole); ok {
					for _, itemRole := range v.UserRoles {
						if itemRole.Value == "admin" {  //超级管理员，目前只支持一种后台管理用户
							return true
						}
					}
				}
			
				return false
			},
			//处理不进行授权的逻辑
			Unauthorized: func(c *gin.Context, code int, message string) {
				c.JSON(code, gin.H{
					"code":    code,
					"message": message,
				})
			},
			LoginResponse: func(c *gin.Context, code int, token string, t time.Time) {

				claims, err := ParseToken(token, []byte(SecretKey))
				if nil != err {
					pc.logger.Error("ParseToken Error", zap.Error(err))
				}
				userName := claims.(jwt.MapClaims)[identityKey].(string)
				pc.logger.Debug("get userName ok", zap.String("userName", userName))

				//将token及expire保存到db
				pc.SaveUserToken(userName, token, t)

				RespData(c, http.StatusOK, code, token)
				// c.JSON(http.StatusOK, gin.H{
				// 	"code":    http.StatusOK,
				// 	"token":   token,
				// 	"expire":  t.Format(time.RFC3339),
				// 	"message": "login successfully",
				// })
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

		auth := r.Group("/v1") //带v1的路由都必须使用Bearer JWT 才能正常访问
		// Refresh time can be longer than token timeout
		auth.GET("/refresh_token", authMiddleware.RefreshHandler)
		auth.Use(authMiddleware.MiddlewareFunc())
		{
			// auth.GET("/hello", helloHandler)
			auth.GET("/user/:id", pc.GetUser)  //根据id获取用户信息
			auth.POST("/chanpassword", pc.ChanPassword) //修改密码
	
		}

	}
}

var ProviderSet = wire.NewSet(NewUsersController, CreateInitControllersFn)

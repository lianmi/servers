package controllers

import (
	"time"
	// "encoding/json"
	"net/http"
	"strconv"

	jwt_v2 "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/lianmi/servers/internal/app/authservice/services"
	"github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/common/codes"
	"github.com/lianmi/servers/internal/pkg/models"
	"go.uber.org/zap"
)

type UsersController struct {
	logger  *zap.Logger
	service services.UsersService
}

func NewUsersController(logger *zap.Logger, s services.UsersService) *UsersController {
	return &UsersController{
		logger:  logger,
		service: s,
	}
}

func (pc *UsersController) GetUser(c *gin.Context) {
	pc.logger.Debug("GetUser start ...")
	ID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	p, err := pc.service.GetUser(ID)
	if err != nil {
		pc.logger.Error("get User by id error", zap.Error(err))
		c.String(http.StatusInternalServerError, "%+v", err)
		return
	}

	c.JSON(http.StatusOK, p)
}

// 用户注册
func (pc *UsersController) Register(c *gin.Context) {
	var user models.User
	code := codes.InvalidParams

	// binding JSON,本质是将request中的Body中的数据按照JSON格式解析到user变量中，必填字段一定要填
	if c.BindJSON(&user) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {

		// roles := jwt.ExtractClaims(c)
		user.CreatedAt = time.Now() //注意，必须要显式

		user.State = 0                 //预审核
		user.Avatar = common.PubAvatar //公共头像

		//检测手机是否已经注册过了
		if pc.service.ExistUserByMobile(user.Mobile) {

			code = codes.ErrExistUser
		}

		//检测校验码是否正确

		//检测是否主设备登录还是从设备登录

		//手机号码还没注册
		if err := pc.service.Register(&user); err == nil {
			code = codes.SUCCESS
		} else {
			pc.logger.Error("Register user error, Mobile is already registered", zap.Error(err))
			code = codes.ERROR
		}

		RespOk(c, http.StatusOK, code)
	}
}

func IsNum(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func IsDigit(data string) bool {
	for _, item := range data {
		if IsNum(string(item)) {
			continue
		} else {
			return false
		}
	}
	return true
}

func (pc *UsersController) GenerateSmsCode(c *gin.Context) {

	code := codes.InvalidParams

	mobile := c.Param("mobile")
	pc.logger.Debug("GenerateSmsCode start ...", zap.String("mobile", mobile))

	//不是手机
	if len(mobile) != 11 {
		pc.logger.Warn("GenerateSmsCode error", zap.Int("len", len(mobile)))

		code = codes.ERROR
		RespOk(c, http.StatusOK, code)
		return
	}

	//不是全数字
	if !IsDigit(mobile) {
		pc.logger.Warn("GenerateSmsCode Is not Digit")
		code = codes.ERROR
		RespOk(c, http.StatusOK, code)
		return
	}

	isOk := pc.service.GenerateSmsCode(mobile)
	if isOk {
		code = codes.SUCCESS
	} else {
		code = codes.ERROR
	}
	RespOk(c, http.StatusOK, code)
}

func (pc *UsersController) ChanPassword(c *gin.Context) {
	return
}

func (pc *UsersController) GetUserRoles(username string) []*models.Role {

	return pc.service.GetUserRoles(username)
}

func (pc *UsersController) CheckUser(username, password string) bool {
	return pc.service.CheckUser(username, password)
}

func (pc *UsersController) SaveUserToken(username, deviceID string, token string, expire time.Time) bool {
	return pc.service.SaveUserToken(username, deviceID, token, expire)
}
func (pc *UsersController) ExistsTokenInRedis(token string) bool {
	return pc.service.ExistsTokenInRedis(token)
}

func (pc *UsersController) SignOut(c *gin.Context) {
	// c.ClientIP
	claims := jwt_v2.ExtractClaims(c)
	userName := claims[common.IdentityKey].(string)
	deviceID := claims["deviceID"].(string)
	token := jwt_v2.GetToken(c)
	pc.logger.Debug("SignOut",
		zap.String("userName", userName),
		zap.String("deviceID", deviceID),
		zap.String("token", token))

	if pc.service.SignOut(token, userName, deviceID) {
		pc.logger.Debug("SignOut  run ok")
	} else {
		pc.logger.Debug("SignOut  run FAILD")

	}

	RespOk(c, http.StatusOK, 200)
}

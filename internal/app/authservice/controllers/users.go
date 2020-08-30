package controllers

import (
	"time"
	// "encoding/json"
	"net/http"
	"strconv"

	"github.com/lianmi/servers/util/conv"

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

type ResetPassword struct {
	Mobile   string `form:"mobile" json:"mobile" binding:"required"` //注册手机
	Password string `json:"password" validate:"required"`            //用户密码，md5加密
	SmsCode  string `json:"smscode" validate:"required"`             //校验码

}

type ChangePassword struct {
	Username    string `json:"username" validate:"username"`     //用户账号s
	OldPassword string `json:"old_password" validate:"required"` //用户密码，md5加密
	NewPassword string `json:"new_password" validate:"required"` //用户密码，md5加密

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

//封号
func (pc *UsersController) BlockUser(c *gin.Context) {

	pc.logger.Debug("BlockUser start ...")
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusNotFound, nil) //404
		return
	}

	p, err := pc.service.BlockUser(username)
	if err != nil {
		pc.logger.Error("Block User by username error", zap.Error(err))
		c.JSON(http.StatusNotFound, nil) //404
		return
	}

	c.JSON(http.StatusOK, p)
}

//解封
func (pc *UsersController) DisBlockUser(c *gin.Context) {
	pc.logger.Debug("DisBlockUser start ...")
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusNotFound, nil) //404
		return
	}

	p, err := pc.service.DisBlockUser(username)
	if err != nil {
		pc.logger.Error("DisBlockUser User by username error", zap.Error(err))
		// c.String(http.StatusInternalServerError, "%+v", err)
		c.JSON(http.StatusNotFound, nil) //404
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
		pc.logger.Debug("注册",
			zap.Int("user.UserType", user.UserType))

		//初始化一些默认值及当期时间
		user.CreatedAt = time.Now().Unix() //注意，必须要unix时间戳
		user.State = 0                     //预审核
		user.Avatar = common.PubAvatar     //公共头像
		user.AllowType = 3                 //用户加好友枚举，默认是3

		//检测手机是数字
		if !conv.IsDigit(user.Mobile) {
			pc.logger.Error("Register user error, Mobile is not digital")
			code = codes.InvalidParams
			RespFail(c, http.StatusBadRequest, code, "Mobile is not digital")
			return
		}

		//检测手机是否已经注册过了
		if pc.service.ExistUserByMobile(user.Mobile) {
			pc.logger.Error("Register user error, Mobile is already registered")
			code = codes.ErrExistMobile
			RespFail(c, http.StatusBadRequest, code, "Mobile is already registered")
			return
		}

		//检测校验码是否正确
		if !pc.service.CheckSmsCode(user.Mobile, user.SmsCode) {
			pc.logger.Error("Register user error, SmsCode is wrong")
			code = codes.InvalidParams
			RespFail(c, http.StatusBadRequest, code, "SmsCode is wrong")
			return
		}

		//是否是商户， 如果是，则商户信息是必填
		if user.UserType == 2 {
			if user.Province == "" || user.County == "" || user.City == "" || user.Street == "" || user.LegalPerson == "" || user.LegalIdentityCard == "" {
				code = codes.InvalidParams
				RespFail(c, http.StatusBadRequest, code, "商户信息必填")
				return
			}
			pc.logger.Debug("商户注册",
				zap.String("Province", user.Province),
				zap.String("County", user.County),
				zap.String("City", user.City),
				zap.String("Street", user.Street),
				zap.String("LegalPerson", user.LegalPerson),
				zap.String("LegalIdentityCard", user.LegalIdentityCard),
			)
		}

		if userName, err := pc.service.Register(&user); err == nil {
			pc.logger.Debug("Register user success", zap.String("userName", userName))
			code = codes.SUCCESS
		} else {
			pc.logger.Error("Register user error", zap.Error(err))
			code = codes.ERROR
			RespFail(c, http.StatusBadRequest, code, "Register user error")
			return
		}
		RespData(c, http.StatusOK, code, user.Username)
	}
}

// 重置密码
func (pc *UsersController) Resetpwd(c *gin.Context) {
	var user models.User
	var rp ResetPassword
	code := codes.InvalidParams

	// binding JSON,本质是将request中的Body中的数据按照JSON格式解析到ResetPassword变量中，必填字段一定要填
	if c.BindJSON(&rp) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {

		//检测手机是数字
		if !conv.IsDigit(rp.Mobile) {
			pc.logger.Error("ResetPassword error, Mobile is not digital")
			code = codes.InvalidParams
			RespFail(c, http.StatusBadRequest, code, "Mobile is not digital")
			return
		}

		//检测手机是否已经注册， 如果未注册，则返回失败
		if !pc.service.ExistUserByMobile(rp.Mobile) {
			pc.logger.Error("ResetPassword error, Mobile is not registered")
			code = codes.ErrExistMobile
			RespFail(c, http.StatusBadRequest, code, "Mobile is not registered")
			return
		}

		//检测校验码是否正确
		if !pc.service.CheckSmsCode(rp.Mobile, rp.SmsCode) {
			pc.logger.Error("ResetPassword error, SmsCode is wrong")
			code = codes.InvalidParams
			RespFail(c, http.StatusBadRequest, code, "SmsCode is wrong")
			return
		}

		if err := pc.service.Resetpwd(rp.Mobile, rp.Password, &user); err == nil {
			pc.logger.Debug("ResetPassword success", zap.String("userName", user.Username))
			code = codes.SUCCESS
		} else {
			pc.logger.Error("ResetPassword error", zap.Error(err))
			code = codes.ERROR
			RespFail(c, http.StatusBadRequest, code, "Reset password error")
			return
		}
		RespData(c, http.StatusOK, code, user.Username)
	}
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
	if !conv.IsDigit(mobile) {
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

	var cp ChangePassword
	if c.BindJSON(&cp) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {
		//修改密码
		if err := pc.service.ChanPassword(cp.Username, cp.OldPassword, cp.NewPassword); err == nil {
			pc.logger.Debug("ChanPassword  run ok")
			RespOk(c, http.StatusOK, 200)
		} else {
			pc.logger.Debug("ChanPassword  run FAILD")
			RespFail(c, http.StatusBadRequest, 400, "修改密码失败")

		}

	}

	return
}

func (pc *UsersController) GetUserRoles(username string) []*models.Role {

	return pc.service.GetUserRoles(username)
}

func (pc *UsersController) CheckUser(isMaster bool, smscode, username, password, deviceID, os string, clientType int) bool {
	return pc.service.CheckUser(isMaster, smscode, username, password, deviceID, os, clientType)
}

func (pc *UsersController) SaveUserToken(username, deviceID string, token string, expire time.Time) bool {
	return pc.service.SaveUserToken(username, deviceID, token, expire)
}
func (pc *UsersController) ExistsTokenInRedis(deviceID, token string) bool {
	return pc.service.ExistsTokenInRedis(deviceID, token)
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

//授权新创建的群组
func (pc *UsersController) ApproveTeam(c *gin.Context) {
	claims := jwt_v2.ExtractClaims(c)
	userName := claims[common.IdentityKey].(string)
	deviceID := claims["deviceID"].(string)
	token := jwt_v2.GetToken(c)
	pc.logger.Debug("ApproveTeam",
		zap.String("userName", userName),
		zap.String("deviceID", deviceID),
		zap.String("token", token))

	//读取
	teamID := c.Param("teamid")
	if teamID == "" {
		c.JSON(http.StatusNotFound, nil) //404
		return
	}
	if err := pc.service.ApproveTeam(teamID); err == nil {
		pc.logger.Debug("ApproveTeam  run ok")
		RespOk(c, http.StatusOK, 200)
	} else {
		pc.logger.Debug("ApproveTeam run FAILD")
		RespFail(c, http.StatusBadRequest, 400, "授权新创建的群组失败")

	}

}

//封禁群组
func (pc *UsersController) BlockTeam(c *gin.Context) {
	//读取
	teamID := c.Param("teamid")
	if teamID == "" {
		c.JSON(http.StatusNotFound, nil) //404
		return
	}
	if err := pc.service.BlockTeam(teamID); err == nil {
		pc.logger.Debug("BlockTeam  run ok")
		RespOk(c, http.StatusOK, 200)
	} else {
		pc.logger.Debug("BlockTeam  run FAILD")
		RespFail(c, http.StatusBadRequest, 400, "封禁群组失败")

	}

}

//解封群组
func (pc *UsersController) DisBlockTeam(c *gin.Context) {
	//读取
	teamID := c.Param("teamid")
	if teamID == "" {
		c.JSON(http.StatusNotFound, nil) //404
		return
	}
	if err := pc.service.DisBlockTeam(teamID); err == nil {
		pc.logger.Debug("DisBlockTeam  run ok")
		RespOk(c, http.StatusOK, 200)
	} else {
		pc.logger.Debug("DisBlockTeam  run FAILD")
		RespFail(c, http.StatusBadRequest, 400, "封禁群组失败")

	}

}

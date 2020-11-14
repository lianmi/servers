/*
这个文件是和前端相关的restful接口-用户模块，/v1/user/....
*/
package controllers

import (
	"net/http"
	"strconv"
	"time"

	Auth "github.com/lianmi/servers/api/proto/auth"
	"github.com/lianmi/servers/util/conv"

	jwt_v2 "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/common/codes"
	"github.com/lianmi/servers/internal/pkg/models"
	// uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

func (pc *LianmiApisController) GetUser(c *gin.Context) {
	pc.logger.Debug("GetUser start ...")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespFail(c, http.StatusBadRequest, 500, "id is wrong number")
		return
	}
	if id <= 0 {
		RespFail(c, http.StatusBadRequest, 500, "id is wrong number")
		return
	}

	p, err := pc.service.GetUser(c, &Auth.UserReq{
		Id: id,
	})
	if err != nil {
		pc.logger.Error("get User by id error", zap.Error(err))
		RespFail(c, http.StatusBadRequest, 500, "Get User by id error")
		return
	}

	RespData(c, http.StatusBadRequest, http.StatusOK, p)

}

// 用户注册
func (pc *LianmiApisController) Register(c *gin.Context) {
	var user models.User
	code := codes.InvalidParams

	// binding JSON,本质是将request中的Body中的数据按照JSON格式解析到user变量中，必填字段一定要填
	if c.BindJSON(&user) != nil {
		pc.logger.Error("Register, binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {
		pc.logger.Debug("注册",
			zap.String("user.Nick", user.Nick),
			zap.String("user.Mobile", user.Mobile),
			zap.String("user.SmsCode", user.SmsCode),
			zap.String("user.ContactPerson", user.ContactPerson),
			zap.Int("user.UserType", user.UserType),
			zap.Int("user.Gender", user.Gender),
			zap.String("user.ReferrerUsername", user.ReferrerUsername),
		)

		//初始化一些默认值及当期时间
		user.CreatedAt = time.Now().UnixNano() / 1e6 //注意，必须要unix时间戳，毫秒
		user.State = 0                               //预审核
		user.Avatar = common.PubAvatar               //公共头像
		user.AllowType = 3                           //用户加好友枚举，默认是3

		//检测手机是数字
		if !conv.IsDigit(user.Mobile) {
			pc.logger.Error("Register user error, Mobile is not digital")
			code = codes.ErrNotDigital
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
			code = codes.ErrWrongSmsCode
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

		//检测推荐人，UI负责将id拼接邀请码，也就是用户账号(id+邀请码)
		if user.ReferrerUsername != "" {
			if !pc.service.ExistUserByName(user.ReferrerUsername) {
				pc.logger.Error("Register user error, ReferrerUsername is not registered")
				code = codes.ErrNotFoundInviter
				RespFail(c, http.StatusBadRequest, code, "ReferrerUsername is not registered")
				return
			}

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
		RespData(c, http.StatusOK, code, Auth.RegisterResp{
			Username: user.Username,
		})
	}
}

// 重置密码
func (pc *LianmiApisController) ResetPassword(c *gin.Context) {
	var user models.User
	var rp Auth.ResetPasswordReq
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
			code = codes.ErrNotRegisterMobile
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

		if err := pc.service.ResetPassword(rp.Mobile, rp.Password, &user); err == nil {
			pc.logger.Debug("ResetPassword success", zap.String("userName", user.Username))
			code = codes.SUCCESS
		} else {
			pc.logger.Error("ResetPassword error", zap.Error(err))
			code = codes.ERROR
			RespFail(c, http.StatusBadRequest, code, "Reset password error")
			return
		}

		RespData(c, http.StatusOK, code, &Auth.ResetPasswordResp{
			Username: user.Username,
		})
	}
}

func (pc *LianmiApisController) GenerateSmsCode(c *gin.Context) {

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

func (pc *LianmiApisController) GetUsernameByMobile(c *gin.Context) {

	code := codes.InvalidParams

	mobile := c.Param("mobile")
	pc.logger.Debug("GetUsernameByMobile start ...", zap.String("mobile", mobile))

	//不是手机
	if len(mobile) != 11 {
		pc.logger.Warn("GetUsernameByMobile error", zap.Int("len", len(mobile)))

		code = codes.ERROR
		RespOk(c, http.StatusOK, code)
		return
	}

	//不是全数字
	if !conv.IsDigit(mobile) {
		pc.logger.Warn("GetUsernameByMobile Is not Digit")
		code = codes.ERROR
		RespOk(c, http.StatusOK, code)
		return
	}

	username, err := pc.service.GetUsernameByMobile(mobile)
	if err != nil {
		code = codes.NONEREGISTER
	} else {
		code = codes.SUCCESS
	}
	RespData(c, http.StatusOK, code, &Auth.ResetPasswordResp{
		Username: username,
	})
}

func (pc *LianmiApisController) ChanPassword(c *gin.Context) {
	var mobile string
	code := codes.InvalidParams

	claims := jwt_v2.ExtractClaims(c)
	userName := claims[common.IdentityKey].(string)
	deviceID := claims["deviceID"].(string)
	token := jwt_v2.GetToken(c)

	pc.logger.Debug("ChanPassword",
		zap.String("userName", userName),
		zap.String("deviceID", deviceID),
		zap.String("token", token))

	var req Auth.ChanPasswordReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {
		if req.Oldpasswd == "" {
			pc.logger.Error("Oldpasswd is empty")
			RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段Oldpasswd")
		}
		if req.Password == "" {
			pc.logger.Error("Password is empty")
			RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段Password")
		}
		if req.SmsCode == "" {
			pc.logger.Error("SmsCode is empty")
			RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段SmsCode")
		}

		//检测校验码是否正确
		if !pc.service.CheckSmsCode(mobile, req.SmsCode) {
			pc.logger.Error("ChanPassword error, SmsCode is wrong")
			code = codes.ErrWrongSmsCode
			RespFail(c, http.StatusBadRequest, code, "SmsCode is wrong")
			return
		}

		//修改密码
		if err := pc.service.ChanPassword(userName, &req); err == nil {
			pc.logger.Debug("ChanPassword  run ok")
			RespOk(c, http.StatusOK, 200)
		} else {
			pc.logger.Debug("ChanPassword  run FAILD")
			RespFail(c, http.StatusBadRequest, 400, "修改密码失败")
		}

	}

	return
}

func (pc *LianmiApisController) GetUserRoles(username string) []*models.Role {

	return pc.service.GetUserRoles(username)
}

func (pc *LianmiApisController) CheckUser(isMaster bool, smscode, username, password, deviceID, os string, clientType int) bool {
	return pc.service.CheckUser(isMaster, smscode, username, password, deviceID, os, clientType)
}

func (pc *LianmiApisController) SaveUserToken(username, deviceID string, token string, expire time.Time) bool {
	return pc.service.SaveUserToken(username, deviceID, token, expire)
}
func (pc *LianmiApisController) ExistsTokenInRedis(deviceID, token string) bool {
	return pc.service.ExistsTokenInRedis(deviceID, token)
}

func (pc *LianmiApisController) SignOut(c *gin.Context) {
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

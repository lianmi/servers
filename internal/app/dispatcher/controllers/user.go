/*
这个文件是和前端相关的restful接口-用户模块，/v1/user/....
*/
package controllers

import (
	"net/http"
	"time"

	Auth "github.com/lianmi/servers/api/proto/auth"
	User "github.com/lianmi/servers/api/proto/user"
	"github.com/lianmi/servers/util/conv"

	jwt_v2 "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/common/codes"
	"github.com/lianmi/servers/internal/pkg/models"
	// uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

//根据用户注册id获取用户资料
func (pc *LianmiApisController) GetUser(c *gin.Context) {
	pc.logger.Debug("Get User start ...")
	username := c.Param("id")

	if username == "" {
		RespFail(c, http.StatusBadRequest, 500, "id is empty")
		return
	}

	user, err := pc.service.GetUser(username)
	if err != nil {
		pc.logger.Error("Get User by id error", zap.Error(err))
		RespFail(c, http.StatusBadRequest, 500, "Get  User by id error")
		return
	}

	RespData(c, http.StatusBadRequest, http.StatusOK, user)

}

//多条件不定参数批量分页获取用户列表
func (pc *LianmiApisController) QueryUsers(c *gin.Context) {
	code := codes.InvalidParams
	pc.logger.Debug("Query Users start ...")
	var req User.QueryUsersReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("Query Users, binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {
		if resp, err := pc.service.QueryUsers(&req); err != nil {
			code = codes.ERROR
			RespFail(c, http.StatusBadRequest, code, "Query users error")
			return
		} else {

			RespData(c, http.StatusBadRequest, http.StatusOK, resp)
		}

	}

}

// 用户注册- 支持普通用户及商户注册
func (pc *LianmiApisController) Register(c *gin.Context) {
	var userReq User.User
	code := codes.InvalidParams

	// binding JSON,本质是将request中的Body中的数据按照JSON格式解析到user变量中，必填字段一定要填
	if c.BindJSON(&userReq) != nil {
		pc.logger.Error("Register, binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {
		pc.logger.Debug("注册",
			zap.String("Nick", userReq.Nick), //呢称
			zap.String("Avatar", userReq.Avatar),
			zap.String("Mobile", userReq.Mobile),                     //手机号
			zap.String("SmsCode", userReq.Smscode),                   //短信校验码
			zap.String("ContactPerson", userReq.ContactPerson),       //联系人
			zap.Int("UserType", int(userReq.UserType)),               //用户类型 1-普通 2-商户
			zap.Int("Gender", int(userReq.Gender)),                   //性别
			zap.String("ReferrerUsername", userReq.ReferrerUsername), //推荐人用户id
		)

		//初始化一些默认值及当期时间
		// userReq.CreatedAt = uint64(time.Now().UnixNano() / 1e6) //注意，必须要unix时间戳，毫秒
		// user.State = 0                                       //预审核
		// user.Avatar = common.PubAvatar                       //公共头像
		// user.AllowType = 3                                   //用户加好友枚举，默认是3

		//检测手机是数字
		if !conv.IsDigit(userReq.Mobile) {
			pc.logger.Error("Register user error, Mobile is not digital")
			code = codes.ErrNotDigital
			RespFail(c, http.StatusBadRequest, code, "Mobile is not digital")
			return
		}

		//检测手机是否已经注册过了
		if pc.service.ExistUserByMobile(userReq.Mobile) {
			pc.logger.Error("Register user error, Mobile is already registered")
			code = codes.ErrExistMobile
			RespFail(c, http.StatusBadRequest, code, "Mobile is already registered")
			return
		}

		//检测校验码是否正确
		if !pc.service.CheckSmsCode(userReq.Mobile, userReq.Smscode) {
			pc.logger.Error("Register user error, SmsCode is wrong")
			code = codes.ErrWrongSmsCode
			RespFail(c, http.StatusBadRequest, code, "SmsCode is wrong")
			return
		}

		//检测推荐人，UI负责将id拼接邀请码，也就是用户账号(id+邀请码)
		if userReq.ReferrerUsername != "" {
			if !pc.service.ExistUserByName(userReq.ReferrerUsername) {
				pc.logger.Error("Register user error, ReferrerUsername is not registered")
				code = codes.ErrNotFoundInviter
				RespFail(c, http.StatusBadRequest, code, "ReferrerUsername is not registered")
				return
			}

		}

		user := models.User{
			Username:         userReq.Username,         //用户注册号，自动生成，字母 + 数字
			Password:         userReq.Passwd,           //用户密码，md5加密
			Nick:             userReq.Nick,             //用户呢称，必填
			Gender:           int(userReq.Gender),      //性别
			Avatar:           userReq.Avatar,           //头像url
			Label:            userReq.Label,            //签名标签
			Mobile:           userReq.Mobile,           //注册手机
			Email:            userReq.Email,            //密保邮件，需要发送校验邮件确认
			AllowType:        3,                        //用户加好友枚举，默认是3
			UserType:         int(userReq.UserType),    //用户类型 1-普通，2-商户
			State:            0,                        //状态 0-普通用户，非VIP 1-付费用户(购买会员) 2-封号
			ContactPerson:    userReq.ContactPerson,    //联系人
			ReferrerUsername: userReq.ReferrerUsername, //推荐人，上线；介绍人, 账号的数字部分，app的推荐码就是用户id的数字
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
	var req Auth.ResetPasswordReq
	code := codes.InvalidParams

	// binding JSON,本质是将request中的Body中的数据按照JSON格式解析到ResetPassword变量中，必填字段一定要填
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {

		pc.logger.Debug("Binding JSON succeed",
			zap.String("Mobile", req.Mobile),
			zap.String("SmsCode", req.SmsCode))

		//检测手机是数字
		if !conv.IsDigit(req.Mobile) {
			pc.logger.Error("Reset Password error, Mobile is not digital")
			code = codes.InvalidParams
			RespFail(c, http.StatusBadRequest, code, "Mobile is not digital")
			return
		}
		//不是手机
		if len(req.Mobile) != 11 {
			pc.logger.Warn("Reset Password error", zap.Int("len", len(req.Mobile)))

			code = codes.InvalidParams
			RespFail(c, http.StatusBadRequest, code, "Mobile is not valid")
			return
		}

		//检测手机是否已经注册， 如果未注册，则返回失败
		if !pc.service.ExistUserByMobile(req.Mobile) {
			pc.logger.Error("Reset Password error, Mobile is not registered")
			code = codes.ErrNotRegisterMobile
			RespFail(c, http.StatusBadRequest, code, "Mobile is not registered")
			return
		}

		pc.logger.Debug("ResetPassword 传参  ",
			zap.String("Mobile", req.Mobile),
			zap.String("SmsCode", req.SmsCode),
			zap.String("Password", req.Password),
		)

		//检测校验码是否正确
		if !pc.service.CheckSmsCode(req.Mobile, req.SmsCode) {
			pc.logger.Error("Reset Password error, SmsCode is wrong")
			code = codes.InvalidParams
			RespFail(c, http.StatusBadRequest, code, "SmsCode is wrong")
			return
		}

		user.Mobile = req.Mobile
		if err := pc.service.ResetPassword(req.Mobile, req.Password, &user); err == nil {
			pc.logger.Debug("Reset Password success", zap.String("userName", user.Username))
			code = codes.SUCCESS
		} else {
			pc.logger.Error("Reset Password error", zap.Error(err))
			code = codes.ERROR
			RespFail(c, http.StatusBadRequest, code, "Reset password error")
			return
		}

		RespData(c, http.StatusOK, code, &Auth.ResetPasswordResp{
			Username: user.Username,
		})
	}
}

//生成短信校验码
func (pc *LianmiApisController) GenerateSmsCode(c *gin.Context) {

	code := codes.InvalidParams

	mobile := c.Param("mobile")
	pc.logger.Debug("GenerateSmsCode start ...", zap.String("mobile", mobile))

	//不是手机
	if len(mobile) != 11 {
		pc.logger.Warn("GenerateSmsCode error", zap.Int("len", len(mobile)))

		code = codes.InvalidParams
		RespFail(c, http.StatusBadRequest, code, "Mobile is not valid")
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

//根据手机号获取注册账号id
func (pc *LianmiApisController) GetUsernameByMobile(c *gin.Context) {

	code := codes.InvalidParams

	mobile := c.Param("mobile")
	pc.logger.Debug("GetUsernameByMobile start ...", zap.String("mobile", mobile))

	//不是手机
	if len(mobile) != 11 {
		pc.logger.Warn("GetUsernameByMobile error", zap.Int("len", len(mobile)))

		code = codes.InvalidParams
		RespFail(c, http.StatusBadRequest, code, "Mobile is not valid")
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

//校验短信验证码
func (pc *LianmiApisController) ValidateCode(c *gin.Context) {

	code := codes.InvalidParams

	var req Auth.ValidCodeReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, code, "参数错误, 缺少必填字段")
	} else {
		if req.Mobile == "" {
			pc.logger.Error("Mobile is empty")
			RespFail(c, http.StatusBadRequest, code, "参数错误, 缺少必填字段Mobile")
		}
		if req.Smscode == "" {
			pc.logger.Error("Smscode is empty")
			RespFail(c, http.StatusBadRequest, code, "参数错误, 缺少必填字段Smscode")
		}

		pc.logger.Debug("ValidateCode",
			zap.String("Mobile", req.Mobile),
			zap.String("Smscode", req.Smscode))

		//检测手机是数字
		if !conv.IsDigit(req.Mobile) {
			pc.logger.Error("ValidateCode error, Mobile is not digital")
			code = codes.InvalidParams
			RespFail(c, http.StatusBadRequest, code, "Mobile is not digital")
			return
		}

		//不是手机
		if len(req.Mobile) != 11 {
			pc.logger.Warn("ValidateCode error", zap.Int("len", len(req.Mobile)))

			code = codes.InvalidParams
			RespFail(c, http.StatusBadRequest, code, "Mobile is not valid")
			return
		}

		//检测校验码是否正确
		if pc.service.CheckSmsCode(req.Mobile, req.Smscode) {
			pc.logger.Debug("ValidateCode, v is valid")
			code = codes.SUCCESS
			RespData(c, http.StatusOK, code, &Auth.ValidCodeRsp{
				Success: true,
				Message: "",
			})

		} else {
			pc.logger.Error("ValidateCode, Smscode is invalid")
			code = codes.SUCCESS
			RespData(c, http.StatusOK, code, &Auth.ValidCodeRsp{
				Success: false,
				Message: "Smscode is invalid",
			})
		}

	}

	return
}

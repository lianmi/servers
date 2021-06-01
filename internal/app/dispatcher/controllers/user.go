/*
这个文件是和前端相关的restful接口-用户模块，/v1/user/....
*/
package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/gomodule/redigo/redis"
	Auth "github.com/lianmi/servers/api/proto/auth"
	User "github.com/lianmi/servers/api/proto/user"
	"github.com/lianmi/servers/util/conv"
	"google.golang.org/protobuf/proto"

	jwt_v2 "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/common/codes"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/lianmi/servers/util/aliyunoss"
	"go.uber.org/zap"
)

//其它用户扫描二维码
func (pc *LianmiApisController) QrcodeDownloadURL(c *gin.Context) {
	pc.logger.Debug("QrcodeDownloadURL..")
	RespData(c, http.StatusOK, 200, LMCommon.QrcodeDownloadURL)
}

//根据用户注册id获取用户资料
func (pc *LianmiApisController) GetUser(c *gin.Context) {
	pc.logger.Debug("Get User start ...")
	username := c.Param("id")

	if username == "" {
		RespData(c, http.StatusOK, 500, "id is empty")
		return
	}

	user, err := pc.service.GetUser(username)
	if err != nil {
		pc.logger.Error("Get User by id error", zap.Error(err))
		RespData(c, http.StatusOK, 500, "Get  User by id error")
		return
	}

	RespData(c, http.StatusOK, 200, user)

}

//根据用户注册id获取用户数据库下载url
func (pc *LianmiApisController) GetUserDb(c *gin.Context) {
	pc.logger.Debug("Get User DB url start ...")
	type UserDBUrl struct {
		Objname string `json:"objname" validate:"required"`
	}
	var req UserDBUrl
	if c.BindJSON(&req) != nil {
		pc.logger.Error("GetUserDb, binding JSON error ")
		RespData(c, http.StatusOK, 400, "参数错误, 缺少必填字段")
	}

	// objname := c.Param("objname")

	if req.Objname == "" {
		RespData(c, http.StatusOK, 500, "objname is empty")
		return
	} else {
		pc.logger.Debug("GetUserDb", zap.String("objname", req.Objname))
	}

	url, err := pc.service.GetUserDb(req.Objname)
	if err != nil {
		pc.logger.Error("Get User DB url by id error", zap.Error(err))
		RespData(c, http.StatusOK, 500, "Get  User DB url by id error")
		return
	}

	RespData(c, http.StatusOK, 200, url)

}

//绑定手机
func (pc *LianmiApisController) UserBindmobile(c *gin.Context) {
	pc.logger.Debug("UserBindmobile start ...")

	username, _, isok := pc.CheckIsUser(c)

	if !isok {
		RespFail(c, http.StatusUnauthorized, 401, "token is fail")
		return
	}

	type Bindmobile struct {
		Mobile  string `json:"mobile" validate:"required"`
		SmsCode string `json:"smscode" validate:"required"`
	}
	var req Bindmobile
	if c.BindJSON(&req) != nil {
		pc.logger.Error("UserBindmobile, binding JSON error ")
		RespData(c, http.StatusOK, 400, "参数错误, 缺少必填字段")
	}

	if req.Mobile == "" {
		RespData(c, http.StatusOK, 500, "Mobile is empty")
		return
	}
	if req.SmsCode == "" {
		RespData(c, http.StatusOK, 500, "SmsCode is empty")
		return
	}

	//检测SmsCode是否正确
	if !pc.service.CheckSmsCode(req.Mobile, req.SmsCode) {
		pc.logger.Error("CheckSmsCode error, SmsCode is wrong")

		RespData(c, http.StatusOK, 500, "SmsCode is wrong")
		return
	}

	pc.logger.Debug("UserBindmobile",
		zap.String("Mobile", req.Mobile),
		zap.String("SmsCode", req.SmsCode),
	)

	err := pc.service.UserBindmobile(username, req.Mobile)
	if err != nil {
		pc.logger.Error("UserBindmobile error", zap.Error(err))
		RespData(c, http.StatusOK, 500, "UserBindmobile error")
		return
	}

	RespOk(c, http.StatusOK, 200)

}

//多条件不定参数批量分页获取用户列表
func (pc *LianmiApisController) QueryUsers(c *gin.Context) {
	code := codes.InvalidParams
	pc.logger.Debug("Query Users start ...")
	var req User.QueryUsersReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("Query Users, binding JSON error ")
		RespData(c, http.StatusOK, 400, "参数错误, 缺少必填字段")
	} else {
		if resp, err := pc.service.QueryUsers(&req); err != nil {
			code = codes.ERROR
			RespData(c, http.StatusOK, code, "Query users error")
			return
		} else {

			RespData(c, http.StatusBadRequest, http.StatusOK, resp)
		}

	}

}

// 用户注册- 支持普通用户及商户注册
func (pc *LianmiApisController) Register(c *gin.Context) {
	var userReq User.User
	var avatar string
	code := codes.InvalidParams

	// binding JSON,本质是将request中的Body中的数据按照JSON格式解析到user变量中，必填字段一定要填
	if c.BindJSON(&userReq) != nil {
		pc.logger.Error("Register, binding JSON error ")
		RespData(c, http.StatusOK, 400, "参数错误, 缺少必填字段")
	} else {
		pc.logger.Debug("注册body部分的字段",
			zap.String("Nick", userReq.Nick),                         //呢称
			zap.String("Password", userReq.Password),                 //密码
			zap.String("Avatar", userReq.Avatar),                     //头像objid
			zap.String("Mobile", userReq.Mobile),                     //手机号
			zap.String("SmsCode", userReq.Smscode),                   //短信校验码
			zap.String("TrueName", userReq.TrueName),                 //实名
			zap.Int("UserType", int(userReq.UserType)),               //用户类型 1-普通 2-商户
			zap.Int("Gender", int(userReq.Gender)),                   //性别
			zap.String("ReferrerUsername", userReq.ReferrerUsername), //推荐人
		)

		//检测手机是数字
		if !conv.IsDigit(userReq.Mobile) {
			pc.logger.Error("Register user error, Mobile is not digital")
			code = codes.ErrNotDigital
			RespData(c, http.StatusOK, code, nil)
			return
		}

		//检测手机是否已经注册过了
		if pc.service.ExistUserByMobile(userReq.Mobile) {
			pc.logger.Error("Register user error, Mobile is already registered")
			code = codes.ErrExistMobile
			RespData(c, http.StatusOK, code, nil)
			return
		}

		//检测校验码是否正确
		if !pc.service.CheckSmsCode(userReq.Mobile, userReq.Smscode) {
			pc.logger.Error("Register user error, SmsCode is wrong")
			code = codes.ErrWrongSmsCode
			RespData(c, http.StatusOK, code, nil)
			return
		}

		//检测推荐人，UI负责将id拼接邀请码，也就是用户账号(id+邀请码)
		if userReq.ReferrerUsername != "" {
			if !pc.service.ExistUserByName(userReq.ReferrerUsername) {
				pc.logger.Error("Register user error, ReferrerUsername is not registered")
				code = codes.ErrNotFoundInviter
				RespData(c, http.StatusOK, code, nil)
				return
			}

		}

		// 如果不传头像，则用默认的
		if userReq.Avatar == "" {
			avatar = LMCommon.PubAvatar
		} else {
			avatar = userReq.Avatar //oss  objid形式的字符串, 首个字母不是/
		}
		user := models.User{
			UserBase: models.UserBase{
				Username:         userReq.Username,         //用户注册号，自动生成，字母 + 数字
				Password:         userReq.Password,         //用户密码，md5加密
				Nick:             userReq.Nick,             //用户呢称，必填
				Gender:           int(userReq.Gender),      //性别
				Avatar:           avatar,                   //头像url
				Label:            userReq.Label,            //签名标签
				Mobile:           userReq.Mobile,           //注册手机
				Email:            userReq.Email,            //密保邮件，需要发送校验邮件确认
				AllowType:        3,                        //用户加好友枚举，默认是3
				UserType:         int(userReq.UserType),    //用户类型 1-普通，2-商户
				State:            0,                        //状态 0-普通用户，非VIP 1-付费用户(购买会员) 2-封号
				TrueName:         userReq.TrueName,         //实名
				ReferrerUsername: userReq.ReferrerUsername, //推荐人，上线；介绍人, 账号的数字部分，app的推荐码就是用户id的数字
			},
		}

		if userName, err := pc.service.Register(&user); err == nil {
			pc.logger.Debug("Register user success", zap.String("userName", userName))
			code = codes.SUCCESS
		} else {
			pc.logger.Error("Register user error", zap.Error(err))
			code = codes.ERROR
			RespData(c, http.StatusOK, code, nil)
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
		RespData(c, http.StatusOK, 400, "参数错误, 缺少必填字段")
	} else {

		pc.logger.Debug("Binding JSON succeed",
			zap.String("Mobile", req.Mobile),
			zap.String("SmsCode", req.SmsCode))

		//检测手机是数字
		if !conv.IsDigit(req.Mobile) {
			pc.logger.Error("Reset Password error, Mobile is not digital")
			code = codes.InvalidParams
			RespData(c, http.StatusOK, code, "Mobile is not digital")
			return
		}
		//不是手机
		if len(req.Mobile) != 11 {
			pc.logger.Warn("Reset Password error", zap.Int("len", len(req.Mobile)))

			code = codes.InvalidParams
			RespData(c, http.StatusOK, code, "Mobile is not valid")
			return
		}

		//检测手机是否已经注册， 如果未注册，则返回失败
		if !pc.service.ExistUserByMobile(req.Mobile) {
			pc.logger.Error("Reset Password error, Mobile is not registered")
			code = codes.ErrNotRegisterMobile
			RespData(c, http.StatusOK, code, "Mobile is not registered")
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
			RespData(c, http.StatusOK, code, "SmsCode is wrong")
			return
		}

		user.Mobile = req.Mobile
		if err := pc.service.ResetPassword(req.Mobile, req.Password, &user); err == nil {
			pc.logger.Debug("Reset Password success", zap.String("userName", user.Username))
			code = codes.SUCCESS
		} else {
			pc.logger.Error("Reset Password error", zap.Error(err))
			code = codes.ERROR
			RespData(c, http.StatusOK, code, "Reset password error")
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
		RespData(c, http.StatusOK, code, "Mobile is not valid")
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

	if mobile == "" {
		pc.logger.Warn("GetUsernameByMobile error", zap.Int("len", len(mobile)))

		code = codes.InvalidParams
		RespData(c, http.StatusOK, code, "Mobile is empty")
		return
	}

	//不是手机
	if len(mobile) != 11 {
		pc.logger.Warn("GetUsernameByMobile error", zap.Int("len", len(mobile)))

		code = codes.InvalidParams
		RespData(c, http.StatusOK, code, "Mobile is not valid")
		return
	}

	//不是全数字
	if !conv.IsDigit(mobile) {
		pc.logger.Warn("Mobile Is not Digit")
		code = codes.ERROR
		RespData(c, http.StatusOK, code, "Mobile is not Digit")
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

func (pc *LianmiApisController) CheckUser(isMaster bool, username, password, deviceID, os string, userType int) bool {
	isPass, curOnlineDevieID := pc.service.CheckUser(isMaster, username, password, deviceID, os, userType)

	pc.logger.Debug("LianmiApisController:CheckUser", zap.String("curOnlineDevieID", curOnlineDevieID))
	if curOnlineDevieID != "" {
		//向当前主设备发出踢下线消息
		//构造负载数据
		kickedEventRsp := &Auth.KickedEventRsp{
			Reason:  Auth.KickReason_SamePlatformKick,    //不允许同一个帐号在多个主设备同时登录
			TimeTag: uint64(time.Now().UnixNano() / 1e6), //必填，用来对比离线后上次被踢的时间戳
		}
		data, _ := proto.Marshal(kickedEventRsp)

		if err := pc.SendMessagetoNsq(username, curOnlineDevieID, data, 2, 5); err != nil {
			pc.logger.Error("Failed to Send Kicked Msg To current onlinee Device to ProduceChannel", zap.Error(err))
		} else {
			pc.logger.Debug("向当前主设备发出踢下线消息", zap.String("当前主设备curOnlineDevieID", curOnlineDevieID))
		}
	}
	return isPass

}

//  使用手机及短信验证码登录
func (pc *LianmiApisController) LoginBySmscode(username, mobile, smscode, deviceID, os string, userType int) bool {

	isPass, curOnlineDevieID := pc.service.LoginBySmscode(username, mobile, smscode, deviceID, os, userType)

	pc.logger.Debug("LianmiApisController:LoginBySmscode", zap.Bool("isPass", isPass), zap.String("curOnlineDevieID", curOnlineDevieID))
	if curOnlineDevieID != "" {
		//向当前主设备发出踢下线消息
		//构造负载数据
		kickedEventRsp := &Auth.KickedEventRsp{
			Reason:  Auth.KickReason_SamePlatformKick,    //不允许同一个帐号在多个主设备同时登录
			TimeTag: uint64(time.Now().UnixNano() / 1e6), //必填，用来对比离线后上次被踢的时间戳
		}
		data, _ := proto.Marshal(kickedEventRsp)

		if err := pc.SendMessagetoNsq(username, curOnlineDevieID, data, 2, 5); err != nil {
			pc.logger.Error("Failed to Send Kicked Msg To current onlinee Device to ProduceChannel", zap.Error(err))
		} else {
			pc.logger.Debug("向当前主设备发出踢下线消息", zap.String("当前主设备curOnlineDevieID", curOnlineDevieID))
		}
	}
	return isPass

}

func (pc *LianmiApisController) SaveUserToken(username, deviceID string, token string, expire time.Time) bool {
	return pc.service.SaveUserToken(username, deviceID, token, expire)
}
func (pc *LianmiApisController) ExistsTokenInRedis(deviceID, token string) bool {
	return pc.service.ExistsTokenInRedis(deviceID, token)
}

func (pc *LianmiApisController) GetAllDevices(c *gin.Context) {
	claims := jwt_v2.ExtractClaims(c)
	userName := claims[LMCommon.IdentityKey].(string)
	deviceID := claims["deviceID"].(string)
	token := jwt_v2.GetToken(c)

	pc.logger.Debug("GetAllDevices",
		zap.String("userName", userName),
		zap.String("deviceID", deviceID),
		zap.String("token", token))

	device, err := pc.service.GetAllDevices(userName)
	if err != nil {

		pc.logger.Debug("GetAllDevices  run FAILD")

	}

	RespData(c, http.StatusOK, 200, []string{device})
}

func (pc *LianmiApisController) SignOut(c *gin.Context) {
	claims := jwt_v2.ExtractClaims(c)
	userName := claims[LMCommon.IdentityKey].(string)
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
		RespData(c, http.StatusOK, code, "参数错误, 缺少必填字段")
		return
	} else {
		if req.Mobile == "" {
			pc.logger.Error("Mobile is empty")
			RespData(c, http.StatusOK, code, "参数错误, 缺少必填字段Mobile")
			return
		}
		if req.Smscode == "" {
			pc.logger.Error("Smscode is empty")
			RespData(c, http.StatusOK, code, "参数错误, 缺少必填字段Smscode")
			return
		}

		pc.logger.Debug("ValidateCode",
			zap.String("Mobile", req.Mobile),
			zap.String("Smscode", req.Smscode))

		//检测手机是数字
		if !conv.IsDigit(req.Mobile) {
			pc.logger.Error("ValidateCode error, Mobile is not digital")
			code = codes.InvalidParams
			RespData(c, http.StatusOK, code, "Mobile is not digital")
			return
		}

		//不是手机
		if len(req.Mobile) != 11 {
			pc.logger.Warn("ValidateCode error", zap.Int("len", len(req.Mobile)))

			code = codes.InvalidParams
			RespData(c, http.StatusOK, code, "Mobile is not valid")
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

/*
mqtt broker auth 插件初始化
*/
func (pc *LianmiApisController) SetMqttBrokerRedisAuth(deviceID, password string) {

	pc.logger.Info("MqttBrokerRedisAuth start...")

	setdb := redis.DialDatabase(6) //固定为6
	setPasswd := redis.DialPassword("")

	c1, err := redis.Dial("tcp", "redis:6379", setdb, setPasswd)
	if err != nil {
		pc.logger.Error("redis.Dial db 6 failed", zap.Error(err))
	}
	defer c1.Close()

	_, err = redis.String(c1.Do("SET", deviceID, password))
	if err != nil {
		pc.logger.Error("redis SET failed", zap.Error(err))
	} else {
		pc.logger.Info("SetMqttBrokerRedisAuth SET ok")
	}
	_, err = c1.Do("EXPIRE", deviceID, 30*24*3600) //设置有效期30天
	if err != nil {

		pc.logger.Error("EXPIRE failed", zap.Error(err))
	}
}

func (pc *LianmiApisController) UserAuthWechatTokenCode(context *gin.Context) {
	type DataBodyInfo struct {
		WeChatCode string `json:"wechat_code"`
	}

	var datainfo DataBodyInfo
	if err := context.BindJSON(&datainfo); err != nil {
		RespData(context, http.StatusOK, codes.InvalidParams, "请求参数错误")
		return
	}
	//if err := json.Unmarshal(data, &datainfo); err != nil {
	//	RespData(context, http.StatusOK, codes.InvalidParams, "code is empty ")
	//	return
	//}

	wxAuthUrl := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code", LMCommon.WechatLoginAppid, LMCommon.WechatLoginSecret, datainfo.WeChatCode) //  common.BASE_IMSDK_URL

	//useridMd5 := Md5String(&userid)
	inBodyMap := make(map[string]interface{})
	byteBody, _ := json.Marshal(inBodyMap)
	req, err := http.NewRequest("GET", wxAuthUrl, strings.NewReader(string(byteBody)))
	if err != nil {
		pc.logger.Error("AuthLogin_3rd_Wechat", zap.Error(err))
	}
	// 添加自定义请求头
	req.Header.Add("Content-Type", "application/json")

	// 其它请求头配置
	client := &http.Client{
		// 设置客户端属性
	}
	resp, err := client.Do(req)
	if err != nil {
		pc.logger.Error("UserAuthWechatTokenCode client ", zap.Error(err))
		//
		RespData(context, http.StatusOK, codes.ERROR, "wechat client request fail ")
		return
	}

	defer resp.Body.Close()

	type ReapDate struct {
		//Data map[string]interface{} `json:"data"`
		Openid       string `json:"openid"`
		AccessToken  string `json:"access_token"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
		Unionid      string `json:"unionid"`
	}

	if resp.StatusCode == http.StatusOK {
		// 成功
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			pc.logger.Error("read Http Body Fail", zap.Error(err))

		}

		pc.logger.Debug("Register IM user Success ", zap.String("body", string(body)))

		respMap := ReapDate{}

		err = json.Unmarshal(body, &respMap)

		if err != nil {

			pc.logger.Error("wechat client deserial fail ")
			RespData(context, http.StatusOK, codes.ERROR, "wechat register fail ")

			return

		}

		pc.logger.Debug("resp data ", zap.Any("data Map ", respMap))
		//chatroomId = respMap.Data["id"].(string)

		if respMap.Openid == "" {
			pc.logger.Error("wechat client not found openid ")
			RespData(context, http.StatusOK, codes.ERROR, "wechat client not found openid ")
			return
		}
		// 通过openid 判断 是否存在用户
		username, err := pc.service.GetUserByWechatOpenid(respMap.Openid)
		if err != nil {
			pc.logger.Error("用户为注册")
			RespData(context, http.StatusOK, codes.ERROR, "用户未绑定微信")
			return
		}

		//
		// 存在则生成一个 uuid 给用户 , 用于登陆获取上下文
		uuidCache := uuid.NewV4().String()
		cacheUserWxOpenID := fmt.Sprintf("WXToken:%s", uuidCache)
		pc.cacheMap[cacheUserWxOpenID] = username

		RespData(context, http.StatusOK, 200, uuidCache)

	} else {
		pc.logger.Error("鉴权错误")
		RespData(context, http.StatusOK, codes.ERROR, "鉴权错误")
		return
	}

}

//根据微信code获取用户唯一id及信息
func (pc *LianmiApisController) UserBindWechat(weChatCode string) (string, error) {
	var err error
	var username string

	wxAuthUrl := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code", LMCommon.WechatLoginAppid, LMCommon.WechatLoginSecret, weChatCode) //  common.BASE_IMSDK_URL

	//useridMd5 := Md5String(&userid)
	inBodyMap := make(map[string]interface{})
	byteBody, _ := json.Marshal(inBodyMap)
	req, err := http.NewRequest("GET", wxAuthUrl, strings.NewReader(string(byteBody)))
	if err != nil {
		pc.logger.Error("AuthLogin_3rd_Wechat", zap.Error(err))
	}
	// 添加自定义请求头
	req.Header.Add("Content-Type", "application/json")
	//

	// 其它请求头配置
	client := &http.Client{
		// 设置客户端属性
	}
	resp, err := client.Do(req)
	if err != nil {
		pc.logger.Error("UserAuthWechatTokenCode client ", zap.Error(err))

		return "", err
	}

	defer resp.Body.Close()

	type ReapDate struct {
		//Data map[string]interface{} `json:"data"`
		Openid       string `json:"openid"`
		AccessToken  string `json:"access_token"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
		Unionid      string `json:"unionid"`
	}
	if resp.StatusCode == http.StatusOK {
		// 成功
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			pc.logger.Error("read Http Body Fail", zap.Error(err))

		}

		pc.logger.Debug("Register IM user Success ", zap.String("body", string(body)))

		respMap := ReapDate{}

		err = json.Unmarshal(body, &respMap)

		if err != nil {

			pc.logger.Error("wechat client deserial fail ")
			return "", err

		}

		pc.logger.Debug("resp data ", zap.Any("data Map ", respMap))
		//chatroomId = respMap.Data["id"].(string)

		if respMap.Openid == "" {
			pc.logger.Error("wechat client not found openid ")
			return "", err
		}

		// 通过openid 判断 是否存在用户
		username, err = pc.service.GetUserByWechatOpenid(respMap.Openid)
		if err != nil {
			// 可以绑定
			// 获取用户微信信息
			var nick string
			var avatar string
			var gender int
			var province string
			var city string

			wechatInfo := pc.GetUserinfoFromWechat(respMap.AccessToken, respMap.Openid)
			if wechatInfo == nil {
				pc.logger.Warn("获取微信信息失败 , 但不影响流程")
			} else {
				nick = wechatInfo.Usernick
				gender = wechatInfo.Sex
				province = wechatInfo.Province
				city = wechatInfo.City

				pc.logger.Debug("wechatInfo",
					zap.String("nick", wechatInfo.Usernick),
					zap.String("nick", wechatInfo.Usernick),
					zap.Int("gender", wechatInfo.Sex),
					zap.String("province", wechatInfo.Province),
					zap.String("city", wechatInfo.City),
				)

				//TODO 利用http下载此用户的头像文件，并上传到阿里云
				avatar, err = pc.DownloadWechatHeadImage(respMap.Openid, wechatInfo.Avator)
				if err != nil {
					pc.logger.Warn("利用http下载此用户的头像文件，并上传到阿里云失败 , 但不影响流程")
				}
			}

			//将用户注册
			user := models.User{
				UserBase: models.UserBase{
					Nick:      nick, //呢称
					Gender:    gender,
					Avatar:    avatar, //头像
					Province:  province,
					City:      city,
					Mobile:    "",             //注册手机
					AllowType: 3,              //用户加好友枚举，默认是3
					UserType:  1,              //用户类型 1-普通，2-商户
					State:     0,              //状态 0-普通用户，非VIP 1-付费用户(购买会员) 2-封号
					WXOpenID:  respMap.Openid, //微信用户唯一id
				},
			}

			if username, err = pc.service.Register(&user); err == nil {
				pc.logger.Debug("Register user success", zap.String("username", username))

			} else {
				pc.logger.Error("Register user error", zap.Error(err))
			}
			return username, nil
		} else {
			// 已经绑定过
			pc.logger.Debug("Register open id 已经绑定过",
				zap.String("openid", respMap.Openid),
				zap.String("username", username),
			)
			return username, nil
		}

		// errBind := pc.service.UpdateUserWxOpenID(username, respMap.Openid)

		// if errBind != nil {
		// 	pc.logger.Error("绑定失败 ", zap.Error(err))
		// 	return "", err
		// } else {
		// 	return username, nil
		// }

	} else {
		pc.logger.Error("鉴权错误")
		return "", err
	}

}

//根据微信头像url下载并上传到阿里云oss
func (pc *LianmiApisController) DownloadWechatHeadImage(openid, avatorUrl string) (string, error) {
	//解析url
	uri, err := url.ParseRequestURI(avatorUrl)
	if err != nil {
		pc.logger.Error("网址错误", zap.Error(err))
		return "", err
	}

	filename := path.Base(uri.Path)
	pc.logger.Debug("[*] Filename " + filename)

	res, err := http.Get(avatorUrl)
	if err != nil {
		pc.logger.Error("http.Get错误", zap.Error(err))
		return "", err
	}
	f, err := os.Create("/tmp/" + filename + ".jpg")
	if err != nil {
		pc.logger.Error("os.Create错误", zap.Error(err))
		return "", err
	}
	io.Copy(f, res.Body)

	// TODO 上传到阿里云
	objname, err := aliyunoss.UploadOssFile("avatars", openid, "/tmp/"+filename+".jpg")
	if err != nil {
		pc.logger.Error("aliyunoss.UploadOssFile错误", zap.Error(err))
		return "", err
	}
	pc.logger.Debug("上传到阿里云成功", zap.String("objname", objname))

	return objname, nil
}

// 通过 token 和openid 获取 微信信息
// 如果获取失败 当会 nil , 否则 有数据
func (pc *LianmiApisController) GetUserinfoFromWechat(token, openid string) *models.WechatBaseInfoDataType {
	wxAuthUrl := fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s", token, openid) //  common.BASE_IMSDK_URL
	//useridMd5 := Md5String(&userid)
	inBodyMap := make(map[string]interface{})
	byteBody, _ := json.Marshal(inBodyMap)
	req, err := http.NewRequest("GET", wxAuthUrl, strings.NewReader(string(byteBody)))
	if err != nil {
		pc.logger.Error("GetUserinfoFromWechat", zap.Error(err))
		return nil
	}
	// 添加自定义请求头
	req.Header.Add("Content-Type", "application/json")
	//

	// 其它请求头配置
	client := &http.Client{
		// 设置客户端属性
	}
	resp, err := client.Do(req)
	if err != nil {
		pc.logger.Error("GetUserinfoFromWechat client ", zap.Error(err))
		return nil
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("success")
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			pc.logger.Error("GetUserinfoFromWechat read Http Body Fail", zap.Error(err))
			return nil
		}

		respMap := models.WechatBaseInfoDataType{}

		err = json.Unmarshal(body, &respMap)
		if err != nil {

			pc.logger.Error("GetUserinfoFromWechat client deserial fail ")
			return nil
		}

		// if respMap.OpenID == openid {
		// 	pc.logger.Error("GetUserinfoFromWechat client not found openid ")
		// 	return nil
		// }

		return &respMap

	} else {
		// 失败
		//fmt.Println("fail")
		pc.logger.Error("GetUserinfoFromWechat fail ")
		return nil
	}
}

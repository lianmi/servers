package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	Global "github.com/lianmi/servers/api/proto/global"
	Order "github.com/lianmi/servers/api/proto/order"
	Service "github.com/lianmi/servers/api/proto/service"
	User "github.com/lianmi/servers/api/proto/user"
	"github.com/lianmi/servers/util/conv"

	jwt_v2 "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/lianmi/servers/internal/app/authservice/services"
	"github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/common/codes"
	"github.com/lianmi/servers/internal/pkg/models"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

type LianmiApisController struct {
	logger  *zap.Logger
	service services.LianmiApisService
}

// type ResetPassword struct {
// 	Mobile   string `form:"mobile" json:"mobile" binding:"required"` //注册手机
// 	Password string `json:"password" validate:"required"`            //用户密码，md5加密
// 	SmsCode  string `json:"smscode" validate:"required"`             //校验码

// }

func NewLianmiApisController(logger *zap.Logger, s services.LianmiApisService) *LianmiApisController {
	return &LianmiApisController{
		logger:  logger,
		service: s,
	}
}

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

	p, err := pc.service.GetUser(id)
	if err != nil {
		pc.logger.Error("get User by id error", zap.Error(err))
		RespFail(c, http.StatusBadRequest, 500, "Get User by id error")
		return
	}

	RespData(c, http.StatusBadRequest, http.StatusOK, &Service.UserRsp{
		User: &User.User{
			Username:          p.Username,
			Gender:            User.Gender(p.Gender),
			Nick:              p.Nick,
			Avatar:            p.Avatar,
			Label:             p.Label,
			Mobile:            p.Mobile,
			Email:             p.Email,
			UserType:          User.UserType(p.UserType),
			State:             User.UserState(p.State),
			Extend:            p.Extend,
			ContactPerson:     p.ContactPerson,
			Introductory:      p.Introductory,
			Province:          p.Province,
			City:              p.City,
			County:            p.County,
			Street:            p.Street,
			Address:           p.Address,
			Branchesname:      p.Branchesname,
			LegalPerson:       p.LegalPerson,
			LegalIdentityCard: p.LegalIdentityCard,
			CreatedAt:         uint64(p.CreatedAt),
			UpdatedAt:         uint64(p.UpdatedAt),
		},
	})
}

//封号
func (pc *LianmiApisController) BlockUser(c *gin.Context) {

	pc.logger.Debug("BlockUser start ...")
	username := c.Param("username")
	if username == "" {
		RespFail(c, http.StatusNotFound, 404, "Param is  empty")
		return
	}

	err := pc.service.BlockUser(username)
	if err != nil {
		pc.logger.Error("Block User by username error", zap.Error(err))
		RespFail(c, http.StatusBadRequest, 500, "Block User by username error")

		return
	}

	RespOk(c, http.StatusOK, 200)
}

//解封
func (pc *LianmiApisController) DisBlockUser(c *gin.Context) {
	pc.logger.Debug("DisBlockUser start ...")
	username := c.Param("username")
	if username == "" {
		RespFail(c, http.StatusNotFound, 404, "Param is empty")
		return
	}

	_, err := pc.service.DisBlockUser(username)
	if err != nil {
		pc.logger.Error("DisBlockUser User by username error", zap.Error(err))
		RespFail(c, http.StatusNotFound, 404, "username  not found")
		return
	}

	RespOk(c, http.StatusOK, 200)
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
		RespData(c, http.StatusOK, code, Service.RegisterResp{
			Username: user.Username,
		})
	}
}

// 重置密码
func (pc *LianmiApisController) ResetPassword(c *gin.Context) {
	var user models.User
	var rp Service.ResetPasswordReq
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

		RespData(c, http.StatusOK, code, &Service.ResetPasswordResp{
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

	var req Service.ChanPasswordReq
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

//授权新创建的群组
func (pc *LianmiApisController) ApproveTeam(c *gin.Context) {
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
		RespFail(c, http.StatusBadRequest, 400, "Param is empty")
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
func (pc *LianmiApisController) BlockTeam(c *gin.Context) {
	//读取
	teamID := c.Param("teamid")
	if teamID == "" {
		RespFail(c, http.StatusBadRequest, 400, "Param is empty")
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
func (pc *LianmiApisController) DisBlockTeam(c *gin.Context) {
	//读取
	teamID := c.Param("teamid")
	if teamID == "" {

		RespFail(c, http.StatusBadRequest, 400, "Param is empty")
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

//增加通用商品
func (pc *LianmiApisController) AddGeneralProduct(c *gin.Context) {
	var og Order.GeneralProduct
	if c.BindJSON(&og) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {
		//增加

		if err := pc.service.AddGeneralProduct(&models.GeneralProduct{
			ProductID:         uuid.NewV4().String(), //商品ID
			ProductName:       og.ProductName,        //商品名称
			ProductType:       int(og.ProductType),   //商品种类枚举
			ProductDesc:       og.ProductDesc,        //商品详细介绍
			ProductPic1Small:  og.ProductPic1Small,   //商品图片1-小图
			ProductPic1Middle: og.ProductPic1Middle,  //商品图片1-中图
			ProductPic1Large:  og.ProductPic1Large,   //商品图片1-大图
			ProductPic2Small:  og.ProductPic2Small,   //商品图片2-小图
			ProductPic2Middle: og.ProductPic2Middle,  //商品图片2-中图
			ProductPic2Large:  og.ProductPic2Large,   //商品图片2-大图
			ProductPic3Small:  og.ProductPic3Small,   //商品图片3-小图
			ProductPic3Middle: og.ProductPic3Middle,  //商品图片3-中图
			ProductPic3Large:  og.ProductPic3Large,   //商品图片3-大图
			Thumbnail:         og.Thumbnail,          //商品短视频缩略图
			ShortVideo:        og.ShortVideo,         //商品短视频
			AllowCancel:       og.AllowCancel,        //是否允许撤单， 默认是可以，彩票类的不可以

		}); err == nil {
			pc.logger.Debug("AddGeneralProduct  run ok")
			RespOk(c, http.StatusOK, 200)
		} else {
			pc.logger.Warn("AddGeneralProduct  run FAILD")
			RespFail(c, http.StatusBadRequest, 400, "增加通用商品失败")

		}

	}
}

//修改通用商品
func (pc *LianmiApisController) UpdateGeneralProduct(c *gin.Context) {
	var og Order.GeneralProduct
	if c.BindJSON(&og) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {
		//修改
		if og.ProductId == "" {
			pc.logger.Warn("ProductId is empty")
			RespFail(c, http.StatusBadRequest, 400, "修改通用商品失败, ProductId 不能为空")
		}

		if err := pc.service.UpdateGeneralProduct(&models.GeneralProduct{
			ProductID:         og.ProductId,         //商品ID
			ProductName:       og.ProductName,       //商品名称
			ProductType:       int(og.ProductType),  //商品种类枚举
			ProductDesc:       og.ProductDesc,       //商品详细介绍
			ProductPic1Small:  og.ProductPic1Small,  //商品图片1-小图
			ProductPic1Middle: og.ProductPic1Middle, //商品图片1-中图
			ProductPic1Large:  og.ProductPic1Large,  //商品图片1-大图
			ProductPic2Small:  og.ProductPic2Small,  //商品图片2-小图
			ProductPic2Middle: og.ProductPic2Middle, //商品图片2-中图
			ProductPic2Large:  og.ProductPic2Large,  //商品图片2-大图
			ProductPic3Small:  og.ProductPic3Small,  //商品图片3-小图
			ProductPic3Middle: og.ProductPic3Middle, //商品图片3-中图
			ProductPic3Large:  og.ProductPic3Large,  //商品图片3-大图
			Thumbnail:         og.Thumbnail,         //商品短视频缩略图
			ShortVideo:        og.ShortVideo,        //商品短视频
			AllowCancel:       og.AllowCancel,       //是否允许撤单， 默认是可以，彩票类的不可以

		}); err == nil {
			pc.logger.Debug("AddGeneralProduct  run ok")
			RespOk(c, http.StatusOK, 200)
		} else {
			pc.logger.Warn("AddGeneralProduct  run FAILD")
			RespFail(c, http.StatusBadRequest, 400, "修改通用商品失败")

		}

	}
}

func (pc *LianmiApisController) GetGeneralProductByID(c *gin.Context) {
	productId := c.Param("productid")
	if productId == "" {
		RespFail(c, http.StatusBadRequest, 400, "productid is empty")
		return
	}

	p, err := pc.service.GetGeneralProductByID(productId)
	if err != nil {
		pc.logger.Error("get GeneralProduct by productId error", zap.Error(err))
		RespFail(c, http.StatusBadRequest, 5000, "Get GeneralProduct by productId error")
		return
	}

	//TODO
	c.JSON(http.StatusOK, p)
}

func (pc *LianmiApisController) GetGeneralProductPage(c *gin.Context) {
	pageIndex, err := strconv.ParseInt(c.Param("page"), 10, 32)

	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	pageSize, err := strconv.ParseInt(c.Param("pagesize"), 10, 32)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	productType, err := strconv.ParseInt(c.Param("producttype"), 10, 32)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	gpWhere := models.GeneralProduct{ProductType: int(productType)}

	var total uint64
	ps, err := pc.service.GetGeneralProductPage(int(pageIndex), int(pageSize), &total, gpWhere)
	if err != nil {
		pc.logger.Error("GetGeneralProduct Page by ProductType error", zap.Error(err))
		c.String(http.StatusInternalServerError, "%+v", err)
		return
	}

	c.JSON(http.StatusOK, ps)
}

func (pc *LianmiApisController) DeleteGeneralProduct(c *gin.Context) {
	productId := c.Param("productid")
	if productId == "" {
		RespFail(c, http.StatusBadRequest, 404, "productid is empty")
		return
	}
	if pc.service.DeleteGeneralProduct(productId) {

		RespOk(c, http.StatusOK, 200)
	} else {
		RespFail(c, http.StatusBadRequest, 400, "delete GeneralProduct failed")
	}

}

//获取空闲的在线客服id数组
func (pc *LianmiApisController) QueryCustomerServices(c *gin.Context) {
	var req Service.QueryCustomerServiceReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {
		csList, err := pc.service.QueryCustomerServices(&req)

		if err != nil {
			RespFail(c, http.StatusBadRequest, 400, "Query CustomerServices failed")
		} else {
			resp := &Service.QueryCustomerServiceResp{}
			for _, onlineCustomerService := range csList {
				resp.OnlineCustomerServices = append(resp.OnlineCustomerServices, &Service.CustomerServiceInfo{
					Username:   onlineCustomerService.Username,
					JobNumber:  onlineCustomerService.JobNumber,
					Type:       Global.CustomerServiceType(onlineCustomerService.Type),
					Evaluation: onlineCustomerService.Evaluation,
					NickName:   onlineCustomerService.Evaluation,
				})
			}

			RespData(c, http.StatusOK, 200, resp)
		}

	}

}

//增加在线客服id
func (pc *LianmiApisController) AddCustomerService(c *gin.Context) {

	var req Service.AddCustomerServiceReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {

		if !(req.Type == 1 || req.Type == 2) {
			RespFail(c, http.StatusBadRequest, 400, "Type参数错误")
		}
		if req.Username == "" {
			RespFail(c, http.StatusBadRequest, 400, "Username参数错误")
		}
		if req.JobNumber == "" {
			RespFail(c, http.StatusBadRequest, 400, "JobNumber参数错误")
		}
		if req.Evaluation == "" {
			RespFail(c, http.StatusBadRequest, 400, "Evaluation参数错误")
		}
		if req.NickName == "" {
			RespFail(c, http.StatusBadRequest, 400, "NickName参数错误")
		}

		err := pc.service.AddCustomerService(&req)

		if err != nil {
			RespFail(c, http.StatusBadRequest, 400, "Add CustomerService failed")
		} else {
			RespOk(c, http.StatusOK, 200)
		}
	}

}

//删除在线客服id
func (pc *LianmiApisController) DeleteCustomerService(c *gin.Context) {

	var req Service.DeleteCustomerServiceReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {

		if req.Username == "" {
			RespFail(c, http.StatusBadRequest, 400, "Username参数错误")
		}

		if pc.service.DeleteCustomerService(&req) == false {
			RespFail(c, http.StatusBadRequest, 400, "Delete CustomerServices failed")
		} else {
			RespOk(c, http.StatusOK, 200)
		}
	}

}

func (pc *LianmiApisController) UpdateCustomerService(c *gin.Context) {

	var req Service.UpdateCustomerServiceReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {
		if !(req.Type == 1 || req.Type == 2) {
			RespFail(c, http.StatusBadRequest, 400, "Type参数错误")
		}
		if req.Username == "" {
			RespFail(c, http.StatusBadRequest, 400, "Username参数错误")
		}
		if req.JobNumber == "" {
			RespFail(c, http.StatusBadRequest, 400, "JobNumber参数错误")
		}
		if req.Evaluation == "" {
			RespFail(c, http.StatusBadRequest, 400, "Evaluation参数错误")
		}
		if req.NickName == "" {
			RespFail(c, http.StatusBadRequest, 400, "NickName参数错误")
		}
		err := pc.service.UpdateCustomerService(&req)

		if err != nil {
			RespFail(c, http.StatusBadRequest, 400, "Update CustomerServices failed")
		} else {
			RespOk(c, http.StatusOK, 200)
		}
	}

}

func (pc *LianmiApisController) QueryGrades(c *gin.Context) {
	var maps string
	var req Service.GradeReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {
		pageIndex := int(req.Page)
		pageSize := int(req.Limit)
		total := new(uint64)
		if pageIndex == 0 {
			pageIndex = 1
		}
		if pageSize == 0 {
			pageSize = 100
		}

		// GetPages 分页返回数据
		if req.StartAt > 0 && req.EndAt > 0 {
			maps = fmt.Sprintf("created_at >= %d and created_at <= %d", req.StartAt, req.EndAt)
		}
		pfList, err := pc.service.QueryGrades(&req, pageIndex, pageSize, total, maps)

		if err != nil {
			RespFail(c, http.StatusBadRequest, 400, "Query Grades( failed")
		} else {
			pages := Service.GradesPage{
				TotalPage: *total,
				// Grades: pfList,
			}
			for _, grade := range pfList {
				pages.Grades = append(pages.Grades, &Service.GradeInfo{
					Title:                   grade.Title,
					AppUsername:             grade.AppUsername,
					CustomerServiceUsername: grade.CustomerServiceUsername,
					JobNumber:               grade.JobNumber,
					Type:                    Global.CustomerServiceType(grade.Type),
					Evaluation:              grade.Evaluation,
					NickName:                grade.NickName,
					Catalog:                 grade.Catalog,
					Desc:                    grade.Desc,
					GradeNum:                int32(grade.GradeNum),
				})
			}

			RespData(c, http.StatusOK, 200, pages)
		}

	}
}

//客服人员增加求助记录，以便发给用户评分
func (pc *LianmiApisController) AddGrade(c *gin.Context) {
	var req Service.AddGradeReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {

		if req.CustomerServiceUsername == "" {
			RespFail(c, http.StatusBadRequest, 400, "CustomerServiceUsername参数错误")
		}

		title, err := pc.service.AddGrade(&req)

		if err != nil {
			RespFail(c, http.StatusBadRequest, 400, "Add Grade failed")
		} else {

			RespData(c, http.StatusOK, 200, &Service.GradeTitleInfo{
				CustomerServiceUsername: req.CustomerServiceUsername,
				Title:                   title,
			})
		}
	}

}

//用户提交评分
func (pc *LianmiApisController) SubmitGrade(c *gin.Context) {
	var req Service.SubmitGradeReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {

		if req.AppUsername == "" {
			RespFail(c, http.StatusBadRequest, 400, "AppUsername参数错误")
		}

		err := pc.service.SubmitGrade(&req)

		if err != nil {
			RespFail(c, http.StatusBadRequest, 400, "Submit Grade failed")
		} else {

			RespOk(c, http.StatusOK, 200)
		}
	}

}

func (pc *LianmiApisController) GetMembershipCardSaleMode(c *gin.Context) {
	var req Service.MembershipCardSaleModeReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {

		if req.BusinessUsername == "" {
			RespFail(c, http.StatusBadRequest, 400, "BusinessUsername参数错误")
		}
		saleMode, err := pc.service.GetMembershipCardSaleMode(req.BusinessUsername)

		if err != nil {
			RespFail(c, http.StatusBadRequest, 400, "Get Membership Card Sale Mode failed")
		} else {

			RespData(c, http.StatusOK, 200, &Service.MembershipCardSaleModeResp{
				SaleType: Global.MembershipCardSaleType(saleMode),
			})
		}
	}

}

func (pc *LianmiApisController) SetMembershipCardSaleMode(c *gin.Context) {
	var req Service.MembershipCardSaleModeReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {

		if req.BusinessUsername == "" {
			RespFail(c, http.StatusBadRequest, 400, "BusinessUsername参数错误")
		}

		saleType := int(req.SaleType)
		if saleType == 0 {
			saleType = 1
		}

		if !(saleType == 1 || saleType == 2) {
			RespFail(c, http.StatusBadRequest, 400, "Set Membership Card Sale Mode failed")
		}

		err := pc.service.SetMembershipCardSaleMode(req.BusinessUsername, saleType)

		if err != nil {
			RespFail(c, http.StatusBadRequest, 400, "Set Membership Card Sale Mode failed")
		} else {

			RespOk(c, http.StatusOK, 200)
		}
	}

}

func (pc *LianmiApisController) GetBusinessMembership(c *gin.Context) {
	var req Service.GetBusinessMembershipReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {

		resp, err := pc.service.GetBusinessMembership(req.IsRebate)

		if err != nil {
			RespFail(c, http.StatusBadRequest, 400, "Get BusinessMembership failed")
		} else {
			// resp := &Service.GetBusinessMembershipResp{
			// 	Totalmembers: 1,
			// }
			//TODO
			RespData(c, http.StatusOK, 200, resp)
		}
	}

}

//调用此接口前，需要调用10-3 发起转账,
//在本地签名，然后携带签名后的交易数据提交到服务端，返回区块高度，交易哈希
//会员付费， 可以他人代付， 如果他人代付，自动成为其推荐人, 强制归属同一个商户,
//支付成功后，向用户发出通知
//如果用户是自行注册的，提醒用户输入商户的推荐码
func (pc *LianmiApisController) PayForMembership(c *gin.Context) {
	claims := jwt_v2.ExtractClaims(c)
	userName := claims[common.IdentityKey].(string)
	deviceID := claims["deviceID"].(string)
	token := jwt_v2.GetToken(c)

	pc.logger.Debug("PayForMembership",
		zap.String("userName", userName),
		zap.String("deviceID", deviceID),
		zap.String("token", token))

	var req Service.PayForMembershipReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {

		err := pc.service.PayForMembership(req.PayUsername)

		if err != nil {
			RespFail(c, http.StatusBadRequest, 400, "PayForMembership failed")
		} else {
			RespOk(c, http.StatusOK, 200)
		}
	}
}

package controllers

import (
	"time"
	// "encoding/json"
	"net/http"
	"strconv"

	Order "github.com/lianmi/servers/api/proto/order"
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
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	p, err := pc.service.GetUser(id)
	if err != nil {
		pc.logger.Error("get User by id error", zap.Error(err))
		c.String(http.StatusInternalServerError, "%+v", err)
		return
	}

	c.JSON(http.StatusOK, p)
}

//封号
func (pc *LianmiApisController) BlockUser(c *gin.Context) {

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
func (pc *LianmiApisController) DisBlockUser(c *gin.Context) {
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
func (pc *LianmiApisController) Register(c *gin.Context) {
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
		user.CreatedAt = time.Now().UnixNano() / 1e6 //注意，必须要unix时间戳，毫秒
		user.State = 0                               //预审核
		user.Avatar = common.PubAvatar               //公共头像
		user.AllowType = 3                           //用户加好友枚举，默认是3

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
func (pc *LianmiApisController) Resetpwd(c *gin.Context) {
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
func (pc *LianmiApisController) BlockTeam(c *gin.Context) {
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
func (pc *LianmiApisController) DisBlockTeam(c *gin.Context) {
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
			CategoryName:      og.CategoryName,       //商品分类名称
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
			CategoryName:      og.CategoryName,      //商品分类名称
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
		c.String(http.StatusInternalServerError, "%+v", err)
		return
	}

	c.JSON(http.StatusOK, p)
}

/*
&Order.GeneralProduct{
		ProductId:         p.ProductID,
		ProductName:       p.ProductName,       //商品名称
		CategoryName:      p.CategoryName,      //商品分类名称
		ProductDesc:       p.ProductDesc,       //商品详细介绍
		ProductPic1Small:  p.ProductPic1Small,  //商品图片1-小图
		ProductPic1Middle: p.ProductPic1Middle, //商品图片1-中图
		ProductPic1Large:  p.ProductPic1Large,  //商品图片1-大图
		ProductPic2Small:  p.ProductPic2Small,  //商品图片2-小图
		ProductPic2Middle: p.ProductPic2Middle, //商品图片2-中图
		ProductPic2Large:  p.ProductPic2Large,  //商品图片2-大图
		ProductPic3Small:  p.ProductPic3Small,  //商品图片3-小图
		ProductPic3Middle: p.ProductPic3Middle, //商品图片3-中图
		ProductPic3Large:  p.ProductPic3Middle, //商品图片3-大图
		Thumbnail:         p.Thumbnail,         //商品短视频缩略图
		ShortVideo:        p.ShortVideo,        //商品短视频
	}
*/

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
	categoryName := c.Param("categoryname")
	gpWhere := models.GeneralProduct{CategoryName: categoryName}
	var total uint64
	ps, err := pc.service.GetGeneralProductPage(int(pageIndex), int(pageSize), &total, gpWhere)
	if err != nil {
		pc.logger.Error("GetGeneralProduct Page by CategoryName error", zap.Error(err))
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

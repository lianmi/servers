/*
这个文件是和后台相关的restful接口，/admin/....
*/
package controllers

import (
	"strings"
	// "fmt"
	"net/http"
	// "strconv"
	// "time"

	// Global "github.com/lianmi/servers/api/proto/global"
	Auth "github.com/lianmi/servers/api/proto/auth"
	Order "github.com/lianmi/servers/api/proto/order"
	// User "github.com/lianmi/servers/api/proto/user"
	// "github.com/lianmi/servers/util/conv"

	jwt_v2 "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	// "github.com/lianmi/servers/internal/app/dispatcher/services"
	"github.com/lianmi/servers/internal/common"
	// "github.com/lianmi/servers/internal/common/codes"
	"github.com/lianmi/servers/internal/pkg/models"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

//检测用户是否有使用后台接口的权限
func (pc *LianmiApisController) CheckIsAdmin(c *gin.Context) bool {
	claims := jwt_v2.ExtractClaims(c)
	userName := claims[common.IdentityKey].(string)
	// deviceID := claims["deviceID"].(string)
	// token := jwt_v2.GetToken(c)

	// 用户是以admin开头的账号是后台管理账号
	if strings.HasPrefix(userName, "admin") {
		return true
	}

	return false

}

//封号
func (pc *LianmiApisController) BlockUser(c *gin.Context) {
	if !pc.CheckIsAdmin(c) {
		return
	}
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
	if !pc.CheckIsAdmin(c) {
		return
	}

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

//授权新创建的群组
func (pc *LianmiApisController) ApproveTeam(c *gin.Context) {
	if !pc.CheckIsAdmin(c) {
		return
	}

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
	if !pc.CheckIsAdmin(c) {
		return
	}

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
	if !pc.CheckIsAdmin(c) {
		return
	}

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
	if !pc.CheckIsAdmin(c) {
		return
	}

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
	if !pc.CheckIsAdmin(c) {
		return
	}

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

//删除通用商品
func (pc *LianmiApisController) DeleteGeneralProduct(c *gin.Context) {
	if !pc.CheckIsAdmin(c) {
		return
	}

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

//增加在线客服id
func (pc *LianmiApisController) AddCustomerService(c *gin.Context) {
	if !pc.CheckIsAdmin(c) {
		return
	}

	var req Auth.AddCustomerServiceReq
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
	if !pc.CheckIsAdmin(c) {
		return
	}

	var req Auth.DeleteCustomerServiceReq
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
	if !pc.CheckIsAdmin(c) {
		return
	}

	var req Auth.UpdateCustomerServiceReq
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
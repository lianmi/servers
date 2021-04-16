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

	jwt_v2 "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	// "github.com/lianmi/servers/internal/app/dispatcher/services"
	"github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/common/codes"
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
		RespData(c, http.StatusOK, 404, "Param is empty")
		return
	}

	err := pc.service.BlockUser(username)
	if err != nil {
		pc.logger.Error("Block User by username error", zap.Error(err))
		RespData(c, http.StatusOK, 500, "Block User by username error")

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
		RespData(c, http.StatusOK, 404, "Param is empty")
		return
	}

	err := pc.service.DisBlockUser(username)
	if err != nil {
		pc.logger.Error("DisBlockUser User by username error", zap.Error(err))
		RespData(c, http.StatusOK, 404, "username  not found")
		return
	}

	RespOk(c, http.StatusOK, 200)
}

//后台: 授权新创建的群组
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
		RespData(c, http.StatusOK, 400, "Param is empty")
		return
	}
	if err := pc.service.ApproveTeam(teamID); err == nil {
		pc.logger.Debug("ApproveTeam  run ok")
		RespOk(c, http.StatusOK, 200)
	} else {
		pc.logger.Debug("ApproveTeam run FAILD")
		RespData(c, http.StatusOK, 400, "授权新创建的群组失败")

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
		RespData(c, http.StatusOK, 400, "Param is empty")
		return
	}
	if err := pc.service.BlockTeam(teamID); err == nil {
		pc.logger.Debug("BlockTeam  run ok")
		RespOk(c, http.StatusOK, 200)
	} else {
		pc.logger.Debug("BlockTeam  run FAILD")
		RespData(c, http.StatusOK, 400, "封禁群组失败")

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

		RespData(c, http.StatusOK, 400, "Param is empty")
		return
	}
	if err := pc.service.DisBlockTeam(teamID); err == nil {
		pc.logger.Debug("DisBlockTeam  run ok")
		RespOk(c, http.StatusOK, 200)
	} else {
		pc.logger.Debug("DisBlockTeam  run FAILD")
		RespData(c, http.StatusOK, 400, "封禁群组失败")

	}

}

func (pc *LianmiApisController) GetGeneralProductPage(c *gin.Context) {
	if !pc.CheckIsAdmin(c) {
		return
	}
	code := codes.InvalidParams
	var req Order.GetGeneralProductPageReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespData(c, http.StatusOK, code, "参数错误, 缺少必填字段")
	} else {
		resp, err := pc.service.GetGeneralProductPage(&req)
		if err != nil {
			RespData(c, http.StatusOK, code, "获取店铺商品列表错误")
			return
		}

		RespData(c, http.StatusOK, 200, resp)

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
		RespData(c, http.StatusOK, 400, "参数错误, 缺少必填字段")
	} else {
		//增加
		var productPic1Large, productPic2Large, productPic3Large string
		if len(og.ProductPics) >= 1 {
			productPic1Large = og.ProductPics[0].Large
		}
		if len(og.ProductPics) >= 2 {
			productPic2Large = og.ProductPics[1].Large
		}
		if len(og.ProductPics) >= 3 {
			productPic3Large = og.ProductPics[2].Large
		}

		generalProductInfo := &models.GeneralProductInfo{
			ProductId:        uuid.NewV4().String(), //商品ID
			ProductName:      og.ProductName,        //商品名称
			ProductType:      int(og.ProductType),   //商品种类枚举
			ProductDesc:      og.ProductDesc,        //商品详细介绍
			ProductPic1Large: productPic1Large,      //商品图片1-大图
			ProductPic2Large: productPic2Large,      //商品图片2-大图
			ProductPic3Large: productPic3Large,      //商品图片3-大图
			ShortVideo:       og.ShortVideo,         //商品短视频
			// AllowCancel:      *og.AllowCancel,        //是否允许撤单， 默认是可以，彩票类的不可以

		}

		if len(og.DescPics) >= 1 {
			generalProductInfo.DescPic1 = og.DescPics[0]
		}
		if len(og.DescPics) >= 2 {
			generalProductInfo.DescPic2 = og.DescPics[1]
		}
		if len(og.DescPics) >= 3 {
			generalProductInfo.DescPic3 = og.DescPics[2]
		}
		if len(og.DescPics) >= 4 {
			generalProductInfo.DescPic4 = og.DescPics[3]
		}
		if len(og.DescPics) >= 5 {
			generalProductInfo.DescPic5 = og.DescPics[4]
		}
		if len(og.DescPics) >= 6 {
			generalProductInfo.DescPic6 = og.DescPics[5]
		}

		if err := pc.service.AddGeneralProduct(generalProductInfo); err == nil {
			pc.logger.Debug("AddGeneralProduct run ok")
			// NOTE wujehy 添加成功 , 需要同时清理一下 缓存
			delete(pc.cacheMap, "CacheGetGeneralProjectIDs")
			delete(pc.cacheMap, "CacheGetGeneralProjectLists")
			RespOk(c, http.StatusOK, 200)
		} else {
			pc.logger.Warn("AddGeneralProduct run FAILD")
			RespData(c, http.StatusOK, 400, "增加通用商品失败")
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
		RespData(c, http.StatusOK, 400, "参数错误, 缺少必填字段")
	} else {
		//修改
		if og.ProductId == "" {
			pc.logger.Warn("ProductId is empty")
			RespData(c, http.StatusOK, 400, "修改通用商品失败, ProductId 不能为空")
		}

		if len(og.ProductPics) == 0 {
			pc.logger.Warn("ProductPics length is 0")
			RespData(c, http.StatusOK, 400, "修改通用商品失败, ProductPics length is 0")
		}
		var productPic1Large, productPic2Large, productPic3Large string
		if len(og.ProductPics) >= 1 {
			productPic1Large = og.ProductPics[0].Large
		}
		if len(og.ProductPics) >= 2 {
			productPic2Large = og.ProductPics[1].Large
		}
		if len(og.ProductPics) >= 3 {
			productPic3Large = og.ProductPics[2].Large
		}

		generalProductInfo := &models.GeneralProductInfo{
			ProductId:        uuid.NewV4().String(), //商品ID
			ProductName:      og.ProductName,        //商品名称
			ProductType:      int(og.ProductType),   //商品种类枚举
			ProductDesc:      og.ProductDesc,        //商品详细介绍
			ProductPic1Large: productPic1Large,      //商品图片1-大图
			ProductPic2Large: productPic2Large,      //商品图片2-大图
			ProductPic3Large: productPic3Large,      //商品图片3-大图
			ShortVideo:       og.ShortVideo,         //商品短视频
			// AllowCancel:      *og.AllowCancel,        //是否允许撤单， 默认是可以，彩票类的不可以

		}

		if len(og.DescPics) >= 1 {
			generalProductInfo.DescPic1 = og.DescPics[0]
		}
		if len(og.DescPics) >= 2 {
			generalProductInfo.DescPic2 = og.DescPics[1]
		}
		if len(og.DescPics) >= 3 {
			generalProductInfo.DescPic3 = og.DescPics[2]
		}
		if len(og.DescPics) >= 4 {
			generalProductInfo.DescPic4 = og.DescPics[3]
		}
		if len(og.DescPics) >= 5 {
			generalProductInfo.DescPic5 = og.DescPics[4]
		}
		if len(og.DescPics) >= 6 {
			generalProductInfo.DescPic6 = og.DescPics[5]
		}

		if err := pc.service.UpdateGeneralProduct(generalProductInfo); err == nil {
			pc.logger.Debug("AddGeneralProduct  run ok")
			RespOk(c, http.StatusOK, 200)
		} else {
			pc.logger.Warn("AddGeneralProduct  run FAILD")
			RespData(c, http.StatusOK, 400, "修改通用商品失败")

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
		RespData(c, http.StatusOK, 404, "productid is empty")
		return
	}
	if pc.service.DeleteGeneralProduct(productId) {

		RespOk(c, http.StatusOK, 200)
	} else {
		RespData(c, http.StatusOK, 400, "delete GeneralProduct failed")
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
		RespData(c, http.StatusOK, 400, "参数错误, 缺少必填字段")
	} else {

		if !(req.Type == 1 || req.Type == 2) {
			RespData(c, http.StatusOK, 400, "Type参数错误")
		}
		if req.Username == "" {
			RespData(c, http.StatusOK, 400, "Username参数错误")
		}
		if req.JobNumber == "" {
			RespData(c, http.StatusOK, 400, "JobNumber参数错误")
		}
		if req.Evaluation == "" {
			RespData(c, http.StatusOK, 400, "Evaluation参数错误")
		}
		if req.NickName == "" {
			RespData(c, http.StatusOK, 400, "NickName参数错误")
		}

		err := pc.service.AddCustomerService(&req)

		if err != nil {
			RespData(c, http.StatusOK, 400, "Add CustomerService failed")
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
		RespData(c, http.StatusOK, 400, "参数错误, 缺少必填字段")
	} else {

		if req.Username == "" {
			RespData(c, http.StatusOK, 400, "Username参数错误")
		}

		if pc.service.DeleteCustomerService(&req) == false {
			RespData(c, http.StatusOK, 400, "Delete CustomerServices failed")
		} else {
			RespOk(c, http.StatusOK, 200)
		}
	}

}

//修改在线客服资料
func (pc *LianmiApisController) UpdateCustomerService(c *gin.Context) {
	if !pc.CheckIsAdmin(c) {
		return
	}

	var req Auth.UpdateCustomerServiceReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespData(c, http.StatusOK, 400, "参数错误, 缺少必填字段")
	} else {
		if !(req.Type == 1 || req.Type == 2) {
			RespData(c, http.StatusOK, 400, "Type参数错误")
		}
		if req.Username == "" {
			RespData(c, http.StatusOK, 400, "Username参数错误")
		}
		if req.JobNumber == "" {
			RespData(c, http.StatusOK, 400, "JobNumber参数错误")
		}
		if req.Evaluation == "" {
			RespData(c, http.StatusOK, 400, "Evaluation参数错误")
		}
		if req.NickName == "" {
			RespData(c, http.StatusOK, 400, "NickName参数错误")
		}
		//修改在线客服资料
		err := pc.service.UpdateCustomerService(&req)

		if err != nil {
			RespData(c, http.StatusOK, 400, "Update CustomerServices failed")
		} else {
			RespOk(c, http.StatusOK, 200)
		}
	}

}

//将店铺审核通过
func (pc *LianmiApisController) AuditStore(c *gin.Context) {
	if !pc.CheckIsAdmin(c) {
		return
	}
	var req Auth.AuditStoreReq

	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespData(c, http.StatusOK, 400, "参数错误, 缺少必填字段")
	} else {
		if req.BusinessUsername == "" {
			RespData(c, http.StatusOK, 400, "BusinessUsername参数错误")
		}

		err := pc.service.AuditStore(&req)

		if err != nil {
			RespData(c, http.StatusOK, 400, "AuditStore failed")
		} else {
			RespOk(c, http.StatusOK, 200)
		}
	}
}

/*

func (s *ApiAdapter) SearchPreference(keyword string, page int, pageSize int) (p *[]models.PreferenceItem, err error) {
	// panic("implement me")
	p = new([]models.PreferenceItem)
	keywordStr := fmt.Sprintf("%%%s%%", keyword)
	offset := page*pageSize - pageSize
	currenDB := s.db.Model(&models.PreferenceItem{}).Not(&models.PreferenceItem{Type: 1})
	if keyword != "" {
		currenDB = currenDB.Where("preference_id LIKE ? ", keywordStr)
	}
	err = currenDB.Limit(pageSize).Offset(offset).Find(p).Error
	return
}

*/

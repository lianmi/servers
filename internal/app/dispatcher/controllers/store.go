/*
这个文件是和前端相关的restful接口-用户模块，/v1/store/....
*/
package controllers

import (
	"net/http"
	// "time"

	Order "github.com/lianmi/servers/api/proto/order"
	User "github.com/lianmi/servers/api/proto/user"

	// jwt_v2 "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	// "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/common/codes"
	// "github.com/lianmi/servers/internal/pkg/models"
	// uuid "github.com/satori/go.uuid"
	// "go.uber.org/zap"
)

//根据商户注册id获取店铺资料
func (pc *LianmiApisController) GetStore(c *gin.Context) {
	code := codes.InvalidParams
	businessUsername := c.Param("id")

	if businessUsername == "" {
		RespFail(c, http.StatusBadRequest, 500, "id is empty")
		return
	}

	store, err := pc.service.GetStore(businessUsername)
	if err != nil {
		RespFail(c, http.StatusBadRequest, 500, err.Error())
		return
	}

	code = codes.SUCCESS
	RespData(c, http.StatusOK, code, store)
}

//增加或修改店铺资料
func (pc *LianmiApisController) AddStore(c *gin.Context) {

	code := codes.InvalidParams
	var req User.Store
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, code, "参数错误, 缺少必填字段")
	} else {
		if req.Province == "" || req.County == "" || req.City == "" || req.Street == "" || req.LegalPerson == "" || req.LegalIdentityCard == "" {
			RespFail(c, http.StatusBadRequest, code, "商户地址信息必填")
			return
		}
		if req.BusinessUsername == "" {
			RespFail(c, http.StatusBadRequest, code, "商户注册账号id必填")
			return
		}
		if req.Branchesname == "" {
			RespFail(c, http.StatusBadRequest, code, "商户店铺名称必填")
			return
		}
		if req.BusinessLicenseUrl == "" {
			RespFail(c, http.StatusBadRequest, code, "营业执照url必填")
			return
		}
		if req.Wechat == "" {
			RespFail(c, http.StatusBadRequest, code, "微信必填")
			return
		}
		if req.Longitude == 0.00 {
			RespFail(c, http.StatusBadRequest, code, "商户地址的经度必填")
			return
		}
		if req.Latitude == 0.00 {
			RespFail(c, http.StatusBadRequest, code, "商户地址的纬度必填")
			return
		}

		//保存或增加
		if err := pc.service.AddStore(&req); err != nil {
			RespFail(c, http.StatusBadRequest, code, "保存店铺资料失败")
		} else {
			code = codes.SUCCESS
			RespOk(c, http.StatusOK, code)
		}

	}

}

//修根据gps位置获取一定范围内的店铺列表
func (pc *LianmiApisController) QueryStoresNearby(c *gin.Context) {

	code := codes.InvalidParams
	var req Order.QueryStoresNearbyReq

	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, code, "参数错误, 缺少必填字段")
	} else {

		resp, err := pc.service.GetStores(&req)
		if err != nil {
			RespFail(c, http.StatusBadRequest, code, "获取店铺列表错误")
			return
		}

		RespData(c, http.StatusOK, 200, resp)
	}

}

//获取某个商户的所有商品列表
func (pc *LianmiApisController) ProductsList(c *gin.Context) {

	code := codes.InvalidParams
	var req Order.ProductsListReq

	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, code, "参数错误, 缺少必填字段")
	} else {

		resp, err := pc.service.GetProductsList(&req)
		if err != nil {
			RespFail(c, http.StatusBadRequest, code, "获取店铺商品列表错误")
			return
		}

		RespData(c, http.StatusOK, 200, resp)
	}

}

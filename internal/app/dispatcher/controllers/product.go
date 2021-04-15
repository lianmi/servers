/*
这个文件是和前端相关的restful接口-商品模块，/v1/product/....
*/
package controllers

import (
	"fmt"
	"github.com/lianmi/servers/internal/pkg/models"
	"net/http"

	"github.com/gin-gonic/gin"
	Order "github.com/lianmi/servers/api/proto/order"
	"github.com/lianmi/servers/internal/common/codes"
	"go.uber.org/zap"
)

func (pc *LianmiApisController) GetGeneralProductByID(c *gin.Context) {
	productId := c.Param("productid")
	if productId == "" {
		RespData(c, http.StatusOK, 400, "productid is empty")
		return
	}

	// 从内存读取缓存
	cacheKey := fmt.Sprintf("CacheGetGeneralProductByID:%s", productId)

	productInfo, ok := pc.cacheMap[cacheKey]

	if ok {
		pc.logger.Debug("GetGeneralProductByID 直接内存读取")
		RespData(c, http.StatusOK, 200, productInfo)
		return
	}

	resp, err := pc.service.GetGeneralProductByID(productId)

	if err != nil {
		pc.logger.Error("get GeneralProduct by productId error", zap.Error(err))
		RespData(c, http.StatusOK, 500, "Get GeneralProduct by productId error")
		return
	}
	resp.UpdatedAtInt = resp.UpdatedAt.UnixNano() / 1e6
	pc.cacheMap[cacheKey] = resp
	RespData(c, http.StatusOK, 200, resp)
	return
}

func (pc *LianmiApisController) GetProductInfo(c *gin.Context) {
	productId := c.Param("productid")
	if productId == "" {
		RespData(c, http.StatusOK, 400, "productid is empty")
		return
	}

	resp, err := pc.service.GetProductInfo(productId)
	if err != nil {
		pc.logger.Error("get Product by productId error", zap.Error(err))
		RespData(c, http.StatusOK, 500, "查询商品错误，可能此商品已下架删除")
		return
	}

	RespData(c, http.StatusOK, 200, resp)
}

// 获取该商户的商品列表
func (pc *LianmiApisController) GetStoreProductLists(c *gin.Context) {
	code := codes.InvalidParams
	var req Order.ProductsListReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespData(c, http.StatusOK, code, "参数错误, 缺少必填字段")
	} else {

		resp, err := pc.service.GetStoreProductLists(&req)
		if err != nil {
			RespData(c, http.StatusOK, code, "获取店铺商品列表错误")
			return
		}

		RespData(c, http.StatusOK, 200, resp)
	}
}

//获取某个商户的所有商品列表
func (pc *LianmiApisController) GetProductsList(c *gin.Context) {

	code := codes.InvalidParams
	var req Order.ProductsListReq

	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespData(c, http.StatusOK, code, "参数错误, 缺少必填字段")
	} else {

		resp, err := pc.service.GetProductsList(&req)
		if err != nil {
			RespData(c, http.StatusOK, code, "获取店铺商品列表错误")
			return
		}

		RespData(c, http.StatusOK, 200, resp)
	}

}

//设置商品的子类型
func (pc *LianmiApisController) SetProductSubType(c *gin.Context) {
	code := codes.InvalidParams
	var req Order.ProductSetSubTypeReq

	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespData(c, http.StatusOK, code, "参数错误, 缺少必填字段")
	} else {
		if req.ProductId == "" {
			RespData(c, http.StatusOK, code, "商品ID不能为空")
			return
		}

		err := pc.service.SetProductSubType(&req)
		if err != nil {
			RespData(c, http.StatusOK, code, "设置商品的子类型发生错误")
			return
		}

		RespOk(c, http.StatusOK, 200)
	}

}

// 增加商户支持的彩种
func (pc *LianmiApisController) AdminAddStoreProductItem(context *gin.Context) {
	//
	if !pc.CheckIsAdmin(context) {
		RespFail(context, http.StatusUnauthorized, codes.ErrAuth, "无权访问")
		return
	}

	//type AddStoreProductReq struct {
	//	Store string `json:"store"`
	//	ProductID string `json:"product_id"`
	//}
	code := codes.InvalidParams
	var req models.StoreProductItems

	if context.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespData(context, http.StatusOK, code, "参数错误, 缺少必填字段")
	} else {
		if req.ProductId == "" {
			RespData(context, http.StatusOK, code, "商品ID不能为空")
			return
		}
		if req.StoreUUID == "" {
			RespData(context, http.StatusOK, code, "商户ID不能为空")
			return
		}

		err := pc.service.AddStoreProductItem(&req)
		if err != nil {
			RespData(context, http.StatusOK, code, "设置商户商品失败")
			return
		}

		RespOk(context, http.StatusOK, 200)
	}
}
func (pc *LianmiApisController) GetGeneralProjectIDs(context *gin.Context) {
	//type FindDataReq struct {
	//	Page int `json:"page"`
	//	Limit int `json:"limit"`
	//}
	// 直接返回所有的通用商品的id

	findMap, ok := pc.cacheMap["CacheGetGeneralProjectIDs"]
	if ok {
		// 内存中存在缓存 直接读取内存
		pc.logger.Debug("GetGeneralProjectIDs 从内存直接获取")
		RespData(context, http.StatusOK, codes.SUCCESS, findMap)
		return
	}

	code := codes.InvalidParams
	var req Order.GetGeneralProductPageReq
	req.Limit = 20
	req.Page = 1
	genProductList, err := pc.service.GetGeneralProductFromDB(&req)
	if err != nil {
		RespFail(context, http.StatusUnauthorized, code, "数据查找错误")
		return
	}

	gProductList := make(map[string]int64)

	for index, item := range *genProductList {
		_ = index
		//gProductList = append(gProductList, item.ProductId)
		gProductList[item.ProductId] = item.UpdatedAt.UnixNano() / 1e6
	}
	pc.cacheMap["CacheGetGeneralProjectIDs"] = gProductList
	RespData(context, http.StatusOK, codes.SUCCESS, gProductList)
	return
}

/*
这个文件是和前端相关的restful接口-商品模块，/v1/product/....
*/
package controllers

import (
	"net/http"
	// "strconv"

	"github.com/gin-gonic/gin"
	Order "github.com/lianmi/servers/api/proto/order"
	// "github.com/lianmi/servers/internal/common/codes"
	// "github.com/lianmi/servers/internal/pkg/models"
	"github.com/lianmi/servers/internal/common/codes"
	"go.uber.org/zap"
)

func (pc *LianmiApisController) GetGeneralProductByID(c *gin.Context) {
	productId := c.Param("productid")
	if productId == "" {
		RespFail(c, http.StatusBadRequest, 400, "productid is empty")
		return
	}

	resp, err := pc.service.GetGeneralProductByID(productId)
	if err != nil {
		pc.logger.Error("get GeneralProduct by productId error", zap.Error(err))
		RespFail(c, http.StatusBadRequest, 5000, "Get GeneralProduct by productId error")
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (pc *LianmiApisController) GetProductInfo(c *gin.Context) {
	productId := c.Param("productid")
	if productId == "" {
		RespFail(c, http.StatusBadRequest, 400, "productid is empty")
		return
	}

	resp, err := pc.service.GetProductInfo(productId)
	if err != nil {
		pc.logger.Error("get Product by productId error", zap.Error(err))
		RespFail(c, http.StatusBadRequest, 5000, "Get GeneralProduct by productId error")
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (pc *LianmiApisController) GetGeneralProductPage(c *gin.Context) {
	/*
		var err error
		var gpWhere models.GeneralProduct
		var pageIndex int64
		var pageSize int64
		var productType int64
		pageIndex, err = strconv.ParseInt(c.Param("page"), 10, 32)

		if err != nil {
			// _ = c.AbortWithError(http.StatusBadRequest, err)
			// return
			pageIndex = 1
		}
		pageSize, err = strconv.ParseInt(c.Param("pagesize"), 10, 32)
		if err != nil {
			// _ = c.AbortWithError(http.StatusBadRequest, err)
			// return
			pageSize = 20
		}

		productType, err = strconv.ParseInt(c.Param("producttype"), 10, 32)
		if err != nil {
			// _ = c.AbortWithError(http.StatusBadRequest, err)
			// return
		} else {
			gpWhere = models.GeneralProduct{ProductType: int(productType)}
		}

		var total int64
		ps, err := pc.service.GetGeneralProductPage(int(pageIndex), int(pageSize), &total, gpWhere)
		if err != nil {
			pc.logger.Error("GetGeneralProduct Page by ProductType error", zap.Error(err))
			c.String(http.StatusInternalServerError, "%+v", err)
			return
		}
		c.JSON(http.StatusOK, ps)
	*/
	code := codes.InvalidParams
	var req Order.GetGeneralProductPageReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, code, "参数错误, 缺少必填字段")
	} else {
		resp, err := pc.service.GetGeneralProductPage(&req)
		if err != nil {
			RespFail(c, http.StatusBadRequest, code, "获取店铺商品列表错误")
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


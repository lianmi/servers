/*
这个文件是和前端相关的restful接口-商品模块，/v1/product/....
*/
package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lianmi/servers/internal/pkg/models"
	"go.uber.org/zap"
)

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

	var total int64
	ps, err := pc.service.GetGeneralProductPage(int(pageIndex), int(pageSize), &total, gpWhere)
	if err != nil {
		pc.logger.Error("GetGeneralProduct Page by ProductType error", zap.Error(err))
		c.String(http.StatusInternalServerError, "%+v", err)
		return
	}

	c.JSON(http.StatusOK, ps)
}

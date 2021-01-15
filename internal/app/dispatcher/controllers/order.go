/*
这个文件是和前端相关的restful接口- 订单 模块，/v1/order/....
*/
package controllers

import (
	"github.com/gin-gonic/gin"
	Order "github.com/lianmi/servers/api/proto/order"
	"github.com/lianmi/servers/internal/common/codes"
	"net/http"
)

//商户端: 将完成订单拍照所有图片上链
func (pc *LianmiApisController) UploadOrderImages(c *gin.Context) {
	code := codes.InvalidParams
	var req Order.UploadOrderImagesReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error")
		RespFail(c, http.StatusBadRequest, code, "参数错误, 缺少必填字段")
	} else {
		if req.OrderID == "" {
			pc.logger.Error("OrderID is empty")
			RespFail(c, http.StatusBadRequest, code, "参数错误, 缺少orderID字段")
		}
		if req.Image == "" {
			pc.logger.Error("Image is empty")
			RespFail(c, http.StatusBadRequest, code, "参数错误, 缺少image字段 ")
		}

		resp, err := pc.service.UploadOrderImages(c, &req)
		if err != nil {
			RespFail(c, http.StatusBadRequest, code, "将完成订单拍照所有图片上链时发生错误")
			return
		}

		RespData(c, http.StatusOK, 200, resp)

	}
}

//用户端: 根据 OrderID 获取所有订单拍照图片
func (pc *LianmiApisController) DownloadOrderImage(c *gin.Context) {
	code := codes.InvalidParams
	orderID := c.Param("orderid")
	if orderID == "" {
		RespFail(c, http.StatusBadRequest, 400, "orderid is empty")
		return

	} else {
		resp, err := pc.service.DownloadOrderImage(orderID)
		if err != nil {
			RespFail(c, http.StatusBadRequest, code, "获取所有订单拍照图片时发生错误")
			return
		}

		RespData(c, http.StatusOK, 200, resp)

	}
}

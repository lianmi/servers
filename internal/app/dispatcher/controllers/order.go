/*
这个文件是和前端相关的restful接口- 订单 模块，/v1/order/....
*/
package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	Order "github.com/lianmi/servers/api/proto/order"
	"github.com/lianmi/servers/internal/common/codes"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

//商户端: 将完成订单拍照所有图片上链
func (pc *LianmiApisController) UploadOrderImages(c *gin.Context) {
	code := codes.InvalidParams
	var req Order.UploadOrderImagesReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error")
		RespData(c, http.StatusOK, code, "参数错误, 缺少必填字段")
	} else {
		if req.OrderID == "" {
			pc.logger.Error("OrderID is empty")
			RespData(c, http.StatusOK, code, "参数错误, 缺少orderID字段")
		}
		if req.Image == "" {
			pc.logger.Error("Image is empty")
			RespData(c, http.StatusOK, code, "参数错误, 缺少image字段 ")
		}

		resp, err := pc.service.UploadOrderImages(c, &req)
		if err != nil {
			RespData(c, http.StatusOK, code, "将完成订单拍照所有图片上链时发生错误")
			return
		}

		RespData(c, http.StatusOK, 200, resp)

	}
}

//用户端: 买家将订单body经过RSA加密后提交到彩票中心或第三方公证, mqtt客户端来接收
func (pc *LianmiApisController) UploadOrderBody(c *gin.Context) {
	code := codes.InvalidParams
	var req Order.UploadOrderBodyReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error")
		RespData(c, http.StatusOK, code, "参数错误, 缺少必填字段")
	} else {

		if req.OrderID == "" {
			pc.logger.Error("OrderID is empty")
			RespData(c, http.StatusOK, code, "参数错误, 缺少orderID字段")
		}
		if req.BodyType == 0 {
			pc.logger.Error("BodyType is empty")
			RespData(c, http.StatusOK, code, "参数错误, 缺少BodyType字段 ")
		}
		if req.BodyObjFile == "" {
			pc.logger.Error("BodyObjFile is empty")
			RespData(c, http.StatusOK, code, "参数错误, 缺少BodyObjFile字段 ")
		}

		resp, rsp, err := pc.service.UploadOrderBody(c, &req)
		if err != nil {
			RespData(c, http.StatusOK, code, "买家将订单body经过RSA加密后提交到彩票中心或第三方公证时发生错误")
			return
		}

		//TODO 经过mqtt转发到彩票中心或第三方公证, 需要增加一个事件协议
		//延时1000ms执行

		go func() {
			time.Sleep(1000 * time.Millisecond)
			data, _ := proto.Marshal(rsp)
			if err := pc.SendMessagetoNsq(rsp.NotaryServiceUsername, rsp.NotaryServiceDeviceID, data, 9, 13); err != nil {

				pc.logger.Error("Failed to Send NotaryService(9-13) Msg to ProduceChannel", zap.Error(err))
			} else {
				pc.logger.Debug("向NotaryService发出订单body加密消息(9-13)",
					zap.String("NotaryServiceUsername", rsp.NotaryServiceUsername),
					zap.String("NotaryServiceDeviceID", rsp.NotaryServiceDeviceID),
				)
			}

		}()

		RespData(c, http.StatusOK, 200, resp)

	}
}

//用户端: 根据 OrderID 获取所有订单拍照图片
func (pc *LianmiApisController) DownloadOrderImage(c *gin.Context) {
	code := codes.InvalidParams
	orderID := c.Param("orderid")
	if orderID == "" {
		RespData(c, http.StatusOK, 400, "orderid is empty")
		return

	} else {
		resp, err := pc.service.DownloadOrderImage(orderID)
		if err != nil {
			RespData(c, http.StatusOK, code, "获取所有订单拍照图片时发生错误")
			return
		}

		RespData(c, http.StatusOK, 200, resp)

	}
}

// 用户端: 根据 OrderID 获取此订单在链上的pending状态
func (pc *LianmiApisController) OrderPendingState(c *gin.Context) {
	orderID := c.Param("orderid")
	if orderID == "" {
		RespData(c, http.StatusOK, 400, "orderid is empty")
		return

	} else {
		ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
		resp, err := pc.service.OrderPendingState(ctx, orderID)
		if err != nil {
			RespData(c, http.StatusOK, 500, "获取此订单在链上的pending状态时发生错误")
			return
		}

		RespData(c, http.StatusOK, 200, resp)

	}
}

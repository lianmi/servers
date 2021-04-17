/*
这个文件是和前端相关的restful接口- 订单 模块，/v1/order/....
*/
package controllers

import (
	"context"
	"github.com/lianmi/servers/api/proto/global"
	"github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	uuid "github.com/satori/go.uuid"
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
		RespFail(c, http.StatusOK, code, "参数错误, 缺少必填字段")
	} else {
		if req.OrderID == "" {
			pc.logger.Error("OrderID is empty")
			RespFail(c, http.StatusOK, code, "参数错误, 缺少orderID字段")
		}
		if req.BodyType == 0 {
			pc.logger.Error("BodyType is empty")
			RespFail(c, http.StatusOK, code, "参数错误, 缺少BodyType字段 ")
		}
		if req.BodyObjFile == "" {
			pc.logger.Error("BodyObjFile is empty")
			RespFail(c, http.StatusOK, code, "参数错误, 缺少BodyObjFile字段 ")
		}

		resp, rsp, err := pc.service.UploadOrderBody(c, &req)
		if err != nil {
			RespFail(c, http.StatusOK, code, "买家将订单body经过RSA加密后提交到彩票中心或第三方公证时发生错误")
			return
		}
		if rsp != nil {
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
		}

		RespData(c, http.StatusOK, 200, resp)

	}
}

//用户端: 根据 OrderID 获取所有订单拍照图片
func (pc *LianmiApisController) DownloadOrderImage(c *gin.Context) {
	code := codes.InvalidParams
	orderID := c.Param("orderid")
	if orderID == "" {
		RespFail(c, http.StatusOK, 400, "orderid is empty")
		return

	} else {
		resp, err := pc.service.DownloadOrderImage(orderID)
		if err != nil {
			RespFail(c, http.StatusOK, code, "获取所有订单拍照图片时发生错误")
			return
		}

		RespData(c, http.StatusOK, 200, resp)

	}
}

// 用户端: 根据 OrderID 获取此订单在链上的pending状态
func (pc *LianmiApisController) OrderPendingState(c *gin.Context) {
	orderID := c.Param("orderid")
	if orderID == "" {
		RespFail(c, http.StatusOK, 400, "orderid is empty")
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
func (pc *LianmiApisController) OrderPayToBusiness(context *gin.Context) {
	username, deviceid, isok := pc.CheckIsUser(context)
	_ = deviceid
	_ = username
	if !isok {
		RespFail(context, http.StatusUnauthorized, 401, "token is fail")
		return
	}

	type SendOrderDataTypeReq struct {
		BusinessId string `json:"business_id" binding:"required" `
		ProductId  string `json:"product_id" binding:"required"`
		CouponId   string `json:"coupon_id" `
		Body       string `json:"body" binding:"required"`
		Publickey  string `json:"publickey" binding:"required"`
	}

	req := SendOrderDataTypeReq{}

	if err := context.BindJSON(&req); err != nil {
		RespFail(context, http.StatusOK, codes.InvalidParams, "请求参数错误")
		return
	}

	// 查找商品是否存在
	getProductInfo, err := pc.service.GetGeneralProductByID(req.ProductId)
	if err != nil {
		RespFail(context, http.StatusNotFound, codes.InvalidParams, "商品未找到")
		return
	}

	// 判断商户状态
	getStoreInfo, err := pc.service.GetStore(req.BusinessId)
	if err != nil {
		RespFail(context, http.StatusNotFound, codes.InvalidParams, "商户信息未找到")
		return
	}

	if getStoreInfo.GetStoreType() == global.StoreType_ST_Undefined {
		RespFail(context, http.StatusNotFound, codes.InvalidParams, "商户信息未定义商店类型")
		return
	}

	//pc.logger.Debug("发起订单支付", zap.Int("StoreType", int(getStoreInfo.StoreType)), zap.Int("productType", getProductInfo.ProductType))
	if int(getStoreInfo.GetStoreType()) != getProductInfo.ProductType {
		RespFail(context, http.StatusNotFound, codes.InvalidParams, "商户不支持的商品类型")
		return
	}

	// 创建订单
	orderItem := models.OrderItems{}
	orderItem.OrderId = uuid.NewV4().String()
	orderItem.ProductId = req.ProductId
	orderItem.StoreId = req.BusinessId
	orderItem.UserId = username
	orderItem.Body = req.Body
	orderItem.PublicKey = req.Publickey
	orderItem.OrderStatus = int(global.OrderState_OS_Undefined)
	orderItem.Amounts = getProductInfo.ProductPrice
	orderItem.Fee = common.ChainFee

	// TODO 优惠券处理

	// 入库
	err = pc.service.SavaOrderItemToDB(&orderItem)

	if err != nil {
		pc.logger.Error("订单保存错误", zap.Error(err))
		RespFail(context, http.StatusOK, codes.ERROR, "订单保存错误 , 请重试")
		return
	}

	// 返回支付信息

	// TODO 向 微信发起支付信息码获取
	type RespDataBodyInfo struct {
		BusinessId string  `json:"business_id"`
		ProductId  string  `json:"product_id"`
		Amounts    float64 `json:"amounts"`
		PayCode    string  `json:"pay_code"`
		PayType    int     `json:"pay_type"`
	}
	resp := RespDataBodyInfo{}
	resp.ProductId = orderItem.ProductId
	resp.BusinessId = orderItem.StoreId
	resp.Amounts = orderItem.Amounts + orderItem.Fee
	resp.PayType = 2
	resp.PayCode = "test_weixinzhifucode"
	RespData(context, http.StatusOK, codes.SUCCESS, resp)
	return
}

func (pc *LianmiApisController) OrderCalcPrice(context *gin.Context) {
	username, deviceid, isok := pc.CheckIsUser(context)
	_ = deviceid
	_ = username
	if !isok {
		RespFail(context, http.StatusUnauthorized, 401, "token is fail")
		return
	}

	type SendOrderDataTypeReq struct {
		BusinessId string `json:"business_id" binding:"required" `
		ProductId  string `json:"product_id" binding:"required"`
		CouponId   string `json:"coupon_id" `
		Body       string `json:"body" binding:"required"`
		Publickey  string `json:"publickey" binding:"required"`
	}

	req := SendOrderDataTypeReq{}
	if err := context.BindJSON(&req); err != nil {
		RespFail(context, http.StatusOK, codes.InvalidParams, "请求参数错误")
		return
	}

	// 查找商品是否存在
	getProductInfo, err := pc.service.GetGeneralProductByID(req.ProductId)
	if err != nil {
		RespFail(context, http.StatusNotFound, codes.InvalidParams, "商品未找到")
		return
	}

	// TODO 优惠券处理

	// 返回支付信息

	// TODO 向 微信发起支付信息码获取
	type RespDataBodyInfo struct {
		BusinessId string  `json:"business_id"`
		ProductId  string  `json:"product_id"`
		Amounts    float64 `json:"amounts"`
		//PayCode    string  `json:"pay_code"`
		//PayType    int     `json:"pay_type"`
	}
	resp := RespDataBodyInfo{}
	resp.ProductId = getProductInfo.ProductId
	resp.BusinessId = req.BusinessId
	resp.Amounts = getProductInfo.ProductPrice + common.ChainFee
	//resp.PayType = 2
	//resp.PayCode = "test_weixinzhifucode"
	RespData(context, http.StatusOK, codes.SUCCESS, resp)
	return
}

func (pc *LianmiApisController) OrderGetLists(context *gin.Context) {
	username, deviceid, isok := pc.CheckIsUser(context)
	_ = deviceid
	_ = username
	if !isok {
		RespFail(context, http.StatusUnauthorized, 401, "token is fail")
		return
	}

	type SendOrderDataTypeReq struct {
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	}

	req := SendOrderDataTypeReq{}
	if err := context.BindJSON(&req); err != nil {
		RespFail(context, http.StatusOK, codes.InvalidParams, "请求参数错误")
		return
	}

	if req.Limit < 10 {
		req.Limit = 10
	}

	// 翻页查找 订单信息

	orderList, err := pc.service.GetOrderListByUser(username, req.Limit, req.Offset)
	if err != nil {
		RespFail(context, http.StatusOK, codes.InvalidParams, "未找到订单信息")
		return
	}

	RespData(context, http.StatusOK, codes.SUCCESS, orderList)

}

/*
这个文件是和前端相关的restful接口- 订单 模块，/v1/order/....
*/
package controllers

import (
	"context"
	"fmt"
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

func (pc *LianmiApisController) OrderWechatCallback(context *gin.Context) {
	// 获取订单信息 然后设置支付成功
	type OrderCallbackDataTypeReq struct {
		OrderID string `json:"order_id" binding:"required" `
		Token   string `json:"token" binding:"required"`
	}

	req := OrderCallbackDataTypeReq{}
	if err := context.BindJSON(&req); err != nil {
		RespFail(context, http.StatusOK, codes.InvalidParams, "请求参数错误")
		return
	}

	if req.Token != "lianmi" {
		RespFail(context, http.StatusOK, codes.InvalidParams, "token fail")
		return
	}

	// 查询缓存 当前订单是不是在处理中
	cacheKey := fmt.Sprintf("OrderStatus:%s", req.OrderID)
	orderStatus, isok := pc.cacheMap[cacheKey]
	if isok {
		// 有数据
		orderStatusInt := orderStatus.(int)
		if orderStatusInt != 0 {
			RespFail(context, http.StatusOK, codes.InvalidParams, "订单已在处理中...")
			return
		}
	}

	//通过订单id 查找订单
	orderitem, err := pc.service.GetOrderListByID(req.OrderID)

	if err != nil {
		RespFail(context, http.StatusOK, codes.ERROR, "订单信息错误")
		return
	}

	if orderitem.OrderStatus != int(global.OrderState_OS_Undefined) {
		RespFail(context, http.StatusOK, codes.InvalidParams, "订单已处理")
		return
	}
	// 缓存
	pc.cacheMap[cacheKey] = orderitem.OrderStatus
	orderStatus = orderitem.OrderStatus

	// TODO 一系列处理

	// 将订单设置成 支付完成
	err = pc.service.SetOrderStatusByOrderID(req.OrderID, int(global.OrderState_OS_IsPayed))

	if err != nil {
		//设置订单
		pc.logger.Error("订单设置支付失败", zap.Error(err))
		RespFail(context, http.StatusOK, codes.InvalidParams, "订单处理失败")
		return
	}

	// TODO 发送 支付完成通知

	// TODO 上链 ???

	//pc.service.

	delete(pc.cacheMap, cacheKey)

	RespData(context, http.StatusOK, codes.SUCCESS, "支付成功")
	return
}
func (pc *LianmiApisController) OrderUpdateStatus(context *gin.Context) {
	username, deviceid, isok := pc.CheckIsUser(context)

	_ = deviceid
	if isok {
		RespFail(context, http.StatusUnauthorized, 401, "token is fail")
		return
	}
	// 更新订单状态
	// 设置可以设置的状态
	type OrderCallbackDataTypeReq struct {
		OrderID string `json:"order_id" binding:"required" `
		UserID  string `json:"user_id"`
		StoreID string `json:"store_id"`
		Status  int    `json:"status"`
	}
	req := OrderCallbackDataTypeReq{}
	if err := context.BindJSON(&req); err != nil {
		RespFail(context, http.StatusOK, codes.InvalidParams, "请求参数错误")
		return
	}

	if req.UserID == "" || req.StoreID == "" {
		RespFail(context, http.StatusOK, codes.InvalidParams, "没有指定的对方用户")
		return
	}

	if req.StoreID != "" && req.StoreID != username {
		// 这个是用户端处理的  , 且自己不是商户自己
		//修改订单状态

		err := pc.service.UpdateOrderStatus(username, req.StoreID, req.OrderID, req.Status)
		if err != nil {
			RespFail(context, http.StatusOK, codes.ERROR, "订单状态修改失败")
			return
		}
		RespOk(context, http.StatusOK, codes.SUCCESS)
		return

	}

	if req.UserID != "" && req.UserID != username {
		// 这个是商户端处理的
		err := pc.service.UpdateOrderStatus(req.UserID, username, req.OrderID, req.Status)
		if err != nil {
			RespFail(context, http.StatusOK, codes.ERROR, "订单状态修改失败")
			return
		}
		RespOk(context, http.StatusOK, codes.SUCCESS)
		return
	}
	RespFail(context, http.StatusOK, codes.InvalidParams, "信息错误")
	return
}

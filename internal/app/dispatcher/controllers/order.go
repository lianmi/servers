/*
这个文件是和前端相关的restful接口- 订单 模块，/v1/order/....
*/
package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/iGoogle-ink/gopay"
	"github.com/iGoogle-ink/gopay/wechat"
	User "github.com/lianmi/servers/api/proto/user"

	"github.com/lianmi/servers/api/proto/global"

	Msg "github.com/lianmi/servers/api/proto/msg"
	"github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	uuid "github.com/satori/go.uuid"

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

func (pc *LianmiApisController) OrderPayToBusiness(context *gin.Context) {
	username, _, isok := pc.CheckIsUser(context)

	if !isok {
		RespFail(context, http.StatusUnauthorized, 401, "token is fail")
		return
	}

	type SendOrderDataTypeReq struct {
		BusinessId  string  `json:"business_id" binding:"required" `
		ProductId   string  `json:"product_id" binding:"required"`
		TotalAmount float64 `json:"total_amount"  binding:"required"`
		Fee         float64 `json:"fee"`
		CouponId    string  `json:"coupon_id" `
		Body        string  `json:"body" binding:"required"`
		Publickey   string  `json:"publickey" binding:"required"`
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

	// 生成订单ID
	orderID := uuid.NewV4().String()
	// TODO 优惠券处理

	// 返回支付信息

	// TODO 向 微信发起支付信息码获取
	OutTradeNo := fmt.Sprintf("%s%s%s%s%s", orderID[0:8], orderID[9:13], orderID[14:18], orderID[19:23], orderID[24:36])

	// 获取商户的 商户id
	wxSubMchID := "1608737479" // 这个是特约商户的子商户id
	bm := make(gopay.BodyMap)
	bm.Set("sp_appid", common.WechatPay_appID).
		Set("sp_mchid", common.WechatPay_mchId).
		Set("sub_appid", common.WechatPay_SUBAppid_LM).
		Set("sub_mchid", wxSubMchID). // 这个通过商户id 获取
		Set("out_trade_no", OutTradeNo).
		Set("description", req.ProductId).
		//Set("total_fee", 1).
		//Set("spbill_create_ip", "127.0.0.1").
		Set("notify_url", "https://api.lianmi.cloud/callback/wechat/notify").
		//Set("trade_type", wechat.TradeType_H5).
		//Set("trade_type", wechat).
		//Set("device_info", "APP").
		SetBodyMap("amount", func(bmloc gopay.BodyMap) {
			// 暂时同意 2 毛钱
			bmloc.Set("total", 1).Set("currency", "CNY")
		}).
		//Set("sign_type", wechat.SignTypeRSA).
		SetBodyMap("settle_info", func(bmloc gopay.BodyMap) {
			bmloc.Set("profit_sharing", false)
		})

	pc.logger.Debug("bm", zap.Any("map", bm))

	wxRsp, err := pc.payWechat.V3PartnerTransactionApp(bm)

	if err != nil {
		pc.logger.Error("生成微信支付失败", zap.Error(err))
		RespFail(context, http.StatusOK, codes.ERROR, "生成订单失败, 请重试")
		return
	} else {
		pc.logger.Debug("生成微信预支付成功", zap.String("preid", wxRsp.Response.PrepayId))
	}

	// 生成 支付码
	// 临时转化成 app 的 appid
	pc.payWechat.Appid = common.WechatPay_SUBAppid_LM
	app, err := pc.payWechat.PaySignOfApp(wxRsp.Response.PrepayId)
	if err != nil {
		pc.logger.Error("生成微信支付码失败", zap.Error(err))
		RespFail(context, http.StatusOK, codes.ERROR, "生成支付信息失败,请重试")
		return
	}

	// 创建订单
	orderItem := models.OrderItems{}
	orderItem.OrderId = orderID
	orderItem.ProductId = req.ProductId
	orderItem.StoreId = req.BusinessId
	orderItem.UserId = username
	orderItem.Body = req.Body
	orderItem.PublicKey = req.Publickey
	orderItem.OrderStatus = int(global.OrderState_OS_SendOK) // 设置成 发送
	orderItem.Amounts = req.TotalAmount
	orderItem.Fee = req.Fee

	// 入库
	err = pc.service.SavaOrderItemToDB(&orderItem)

	if err != nil {
		pc.logger.Error("订单保存错误", zap.Error(err))
		RespFail(context, http.StatusOK, codes.ERROR, "订单保存错误 , 请重试")
		return
	}

	type RespDataBodyInfo struct {
		OrderId    string      `json:"order_id"`
		BusinessId string      `json:"business_id"`
		ProductId  string      `json:"product_id"`
		Amounts    float64     `json:"amounts"`
		PayCode    interface{} `json:"pay_code"`
		PayType    int         `json:"pay_type"`
	}

	jsonStr, _ := json.Marshal(app)

	resp := RespDataBodyInfo{}
	resp.OrderId = orderItem.OrderId
	resp.ProductId = orderItem.ProductId
	resp.BusinessId = orderItem.StoreId
	resp.Amounts = orderItem.Amounts
	resp.PayType = 2
	resp.PayCode = string(jsonStr)
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
		//Body       string `json:"body" binding:"required"`
		//Publickey  string `json:"publickey" binding:"required"`
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
		BusinessId string `json:"business_id"`
		ProductId  string `json:"product_id"`
		//Amounts    float64 `json:"amounts"`
		CouponAmount   float64 `json:"coupon_amount"`
		FeeRate        float64 `json:"fee_rate"`
		RateFreeAmount float64 `json:"rate_free_amount"`
		//PayCode    string  `json:"pay_code"`
		//PayType    int     `json:"pay_type"`
	}

	resp := RespDataBodyInfo{}
	resp.ProductId = getProductInfo.ProductId
	resp.BusinessId = req.BusinessId
	resp.FeeRate = common.Rate
	resp.RateFreeAmount = common.RateFreeAmout

	//resp.Amounts = getProductInfo.ProductPrice + common.ChainFee
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
		Limit  int `form:"limit"`
		Offset int `form:"offset"`
		Status int `form:"status"`
	}

	req := SendOrderDataTypeReq{}

	if err := context.BindQuery(&req); err != nil {
		req.Limit = 10
	}

	// 翻页查找 订单信息

	orderList, err := pc.service.GetOrderListByUser(username, req.Limit, req.Offset, req.Status)
	if err != nil {
		RespFail(context, http.StatusOK, codes.InvalidParams, "未找到订单信息")
		return
	}

	RespData(context, http.StatusOK, codes.SUCCESS, orderList)

}

func (pc *LianmiApisController) OrderWechatCallbackRelease(context *gin.Context) {
	//req := wechat.NotifyResponse{}
	// TODO 目前只做订单状态处理 具体校验 暂缓
	pc.logger.Debug("--------微信支付回调 CallbackWalletWeChatNotify---------")
	// notifyReq, err := wechat.V3ParseNotify()

	//var req *http.Request
	//req = context.Request
	req := wechat.NotifyRequest{}
	type RespCallbackDataType struct {
		Code    int    `json:"code" from:"code"`
		Message string `json:"message" from:"err_msg"`
	}
	if err := context.BindJSON(&req); err != nil {
		pc.logger.Error("微信支付请求参数错误", zap.Error(err))
		context.JSON(500, &RespCallbackDataType{Code: 500, Message: "请求参数错误"})
		return
	}
	// 参数处理成功

	if req.Appid != common.WechatPay_appID {
		// 是我们的订单
		context.XML(500, &RespCallbackDataType{Code: 500, Message: "订单appid错误"})
		return
	}

	// 获取订单信息

	orderWechat, err := pc.payWechat.V3PartnerQueryOrder(2, req.SubMchId, req.TransactionId)
	if err != nil {
		// 找不到订单
		context.JSON(404, &RespCallbackDataType{Code: 404, Message: "找不到订单"})
		return
	}

	// 找得到
	pc.logger.Debug("找到支付订单 , ", zap.Any("orderWechat", orderWechat))
	pc.logger.Debug("找到支付订单 , ", zap.Any("id", orderWechat.Response.OutTradeNo))

	// 更性订单状态
	// 订单转化
	OutTradeNo := orderWechat.Response.OutTradeNo
	orderid := fmt.Sprintf("%s-%s-%s-%s-%s", OutTradeNo[0:8], OutTradeNo[8:12], OutTradeNo[12:16], OutTradeNo[16:20], OutTradeNo[20:32])
	// 查询缓存 当前订单是不是在处理中
	cacheKey := fmt.Sprintf("OrderStatus:%s", orderid)
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
	orderitem, err := pc.service.GetOrderListByID(orderid)

	if err != nil {
		RespFail(context, http.StatusOK, codes.ERROR, "订单信息错误")
		return
	}

	if orderitem.OrderStatus != int(global.OrderState_OS_SendOK) {
		RespFail(context, http.StatusOK, codes.InvalidParams, "订单已处理")
		return
	}
	// 缓存
	pc.cacheMap[cacheKey] = orderitem.OrderStatus
	orderStatus = orderitem.OrderStatus

	errChange := pc.service.UpdateOrderStatusByWechatCallback(orderid)
	if errChange != nil {
		pc.logger.Error("更新失败", zap.Error(errChange))
	}

	delete(pc.cacheMap, cacheKey)
	context.JSON(200, &RespCallbackDataType{Code: 200, Message: "SUCCESS"})

	return
	//
	//notification, err := pc.payWechat.V3TransactionQueryOrder(req)
}

// NOTE 这个是手工修改状态的 测试使用
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

	if orderitem.OrderStatus != int(global.OrderState_OS_SendOK) {
		RespFail(context, http.StatusOK, codes.InvalidParams, "订单已处理")
		return
	}
	// 缓存
	pc.cacheMap[cacheKey] = orderitem.OrderStatus
	orderStatus = orderitem.OrderStatus

	// TODO 一系列处理

	// 将订单设置成 支付完成
	err = pc.service.UpdateOrderStatusByWechatCallback(req.OrderID)

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

func (pc *LianmiApisController) OrderGetOrderInfoByID(context *gin.Context) {
	username, deviceid, isok := pc.CheckIsUser(context)

	_ = deviceid
	if !isok {
		RespFail(context, http.StatusUnauthorized, 401, "token is fail")
		return
	}

	orderid := context.Param("id")
	getOrderInfo, err := pc.service.GetOrderListByID(orderid)
	if err != nil {
		RespFail(context, http.StatusNotFound, 404, "未找到数据")
		return
	}

	if getOrderInfo.UserId == username || getOrderInfo.StoreId == username {

	} else {
		RespFail(context, http.StatusNotFound, 404, "未找到数据")
		return
	}

	RespData(context, http.StatusOK, codes.SUCCESS, getOrderInfo)
	return

}

func (pc *LianmiApisController) OrderUpdateStatusByOrderID(context *gin.Context) {
	username, deviceid, isok := pc.CheckIsUser(context)

	_ = deviceid
	if !isok {
		RespFail(context, http.StatusUnauthorized, 401, "token is fail")
		return
	}
	// 更新订单状态
	// 设置可以设置的状态
	type OrderCallbackDataTypeReq struct {
		OrderID string `json:"order_id" binding:"required" `
		Status  int    `json:"status" binding:"required"`
	}

	req := OrderCallbackDataTypeReq{}
	if err := context.BindJSON(&req); err != nil {
		RespFail(context, http.StatusOK, codes.InvalidParams, "请求参数错误")
		return
	}

	// 直接过滤 修改成支付完成状态
	if req.Status == int(global.OrderState_OS_IsPayed) {
		RespFail(context, http.StatusOK, codes.ERROR, "无权修改这个状态")
		return
	}
	// 查询订单

	// 读取用户类型
	//
	userType, err := pc.service.GetUserType(username)
	if err != nil {
		RespFail(context, http.StatusOK, codes.ErrAuth, "用户类型检测异常")
		return
	}

	// 可通过 的修改状态
	if userType == int(User.UserType_Ut_Business) {
		// 商户类型可修改的状态
		if req.Status == int(global.OrderState_OS_Done) ||
			req.Status == int(global.OrderState_OS_Refuse) {
			// 校验通过
		} else {
			RespFail(context, http.StatusOK, codes.ErrAuth, "商户无权修改当前的状态")
			return
		}
	} else if userType == int(User.UserType_Ut_Normal) {
		// 普通用户可以修改的状态
		if req.Status == int(global.OrderState_OS_Confirm) {
			// 校验通过
		} else {
			RespFail(context, http.StatusOK, codes.ErrAuth, "用户无权修改当前的状态")
			return
		}
	} else if userType == int(User.UserType_Ut_Operator) {
		// 管理员直接通过
	} else {
		//
		pc.logger.Error("用户类型检测失败 ", zap.String("userid", username), zap.Int("userTyoe ", userType))
		RespFail(context, http.StatusOK, codes.ErrAuth, "用户类型错误")
		return
	}

	getOrderInfo, err := pc.service.GetOrderListByID(req.OrderID)
	if err != nil {
		RespFail(context, http.StatusOK, codes.InvalidParams, "订单不存在")
		return
	}

	if getOrderInfo.UserId == username || getOrderInfo.StoreId == username {

	} else {
		RespFail(context, http.StatusOK, codes.ERROR, "无权操作这个订单")
		return
	}

	//

	//// 修改订单状态接口
	//// 仅能处理 拒单,接单,确认收获这三种状态
	//// 其他状态均不可以想这个接口处理
	newOrder, err := pc.service.UpdateOrderStatus(getOrderInfo.UserId, getOrderInfo.StoreId, req.OrderID, req.Status)
	_ = newOrder
	if err != nil {
		RespFail(context, http.StatusOK, codes.ERROR, "订单状态修改失败")
		return
	}

	// 订单新状态更新成功 , 可以做其他细化任务
	// 推送用户和商户 变更更时间
	// TODO 如果是 完成 则需要推送到 见证中心
	if req.Status == int(global.OrderState_OS_Confirm) {
		// TODO 推送消息 到 见证中心
	}

	go func() {
		// 推送订单状态变更到商户和用户
		orderChangeReq := new(Msg.RecvMsgEventRsp)
		orderChangeReq.Type = Msg.MessageType_MsgType_Order
		orderChangeReq.Scene = Msg.MessageScene_MsgScene_S2C

		orderProduct := Order.OrderProductBody{}
		orderProduct.OrderID = newOrder.OrderId
		orderProduct.State = global.OrderState(newOrder.OrderStatus)

		orderChangeReq.Body, err = proto.Marshal(&orderProduct)

		if err != nil {
			pc.logger.Error("序列化protobuf order 失败")
			return
		}

		//orderChangeReq.Time

		// 系统到商户
		err1 := pc.nsqClient.BroadcastSystemMsgToAllDevices(orderChangeReq, getOrderInfo.StoreId)
		if err1 != nil {
			pc.logger.Error("推送订单变更事件到用户失败")
		}
		err2 := pc.nsqClient.BroadcastSystemMsgToAllDevices(orderChangeReq, getOrderInfo.UserId)
		if err2 != nil {
			pc.logger.Error("推送订单变更事件到商户失败")

		}
	}()
	RespOk(context, http.StatusOK, codes.SUCCESS)
	return
}

//// 用户或商户手动更改 订单状态
//func (pc *LianmiApisController) OrderUpdateStatus(context *gin.Context) {
//	username, deviceid, isok := pc.CheckIsUser(context)
//
//	_ = deviceid
//	if !isok {
//		RespFail(context, http.StatusUnauthorized, 401, "token is fail")
//		return
//	}
//	// 更新订单状态
//	// 设置可以设置的状态
//	type OrderCallbackDataTypeReq struct {
//		OrderID string `json:"order_id" binding:"required" `
//		UserID  string `json:"user_id"`
//		StoreID string `json:"store_id"`
//		Status  int    `json:"status"`
//	}
//	req := OrderCallbackDataTypeReq{}
//	if err := context.BindJSON(&req); err != nil {
//		RespFail(context, http.StatusOK, codes.InvalidParams, "请求参数错误")
//		return
//	}
//
//	if req.UserID == "" || req.StoreID == "" {
//		RespFail(context, http.StatusOK, codes.InvalidParams, "没有指定的对方用户")
//		return
//	}
//	if req.UserID != "" && req.StoreID != "" {
//		RespFail(context, http.StatusOK, codes.InvalidParams, "没有指定的对方用户")
//		return
//	}
//
//	if req.StoreID != "" && req.StoreID != username {
//		// 这个是用户端处理的  , 且自己不是商户自己
//		//修改订单状态
//
//		newOrder, err := pc.service.UpdateOrderStatus(username, req.StoreID, req.OrderID, req.Status)
//		_ = newOrder
//		if err != nil {
//			RespFail(context, http.StatusOK, codes.ERROR, "订单状态修改失败")
//			return
//		}
//		go func() {
//			// 推送订单状态变更到商户和用户
//
//			orderChangeReq := new(Msg.RecvMsgEventRsp)
//			orderChangeReq.Type = Msg.MessageType_MsgType_Order
//			orderChangeReq.Scene = Msg.MessageScene_MsgScene_S2C
//
//			orderProduct := Order.OrderProductBody{}
//			orderProduct.OrderID = newOrder.OrderId
//			orderProduct.State = global.OrderState(newOrder.OrderStatus)
//
//			orderChangeReq.Body, err = proto.Marshal(&orderProduct)
//
//			if err != nil {
//				pc.logger.Error("序列化protobuf order 失败")
//				return
//			}
//
//			//orderChangeReq.Time
//
//			// 系统到商户
//			err1 := pc.nsqClient.BroadcastSystemMsgToAllDevices(orderChangeReq, username)
//			if err1 != nil {
//				pc.logger.Error("推送订单变更事件到用户失败")
//			}
//			err2 := pc.nsqClient.BroadcastSystemMsgToAllDevices(orderChangeReq, req.StoreID)
//			if err2 != nil {
//				pc.logger.Error("推送订单变更事件到商户失败")
//
//			}
//		}()
//		RespOk(context, http.StatusOK, codes.SUCCESS)
//		return
//
//	}
//
//	if req.UserID != "" && req.UserID != username {
//		// 这个是商户端处理的
//		newOrder, err := pc.service.UpdateOrderStatus(req.UserID, username, req.OrderID, req.Status)
//		_ = newOrder
//
//		if err != nil {
//			RespFail(context, http.StatusOK, codes.ERROR, "订单状态修改失败")
//			return
//		}
//		//TODO  向用户设备推送订单那状态
//
//		go func() {
//			// 推送订单状态变更到商户和用户
//
//			orderChangeReq := new(Msg.RecvMsgEventRsp)
//			orderChangeReq.Type = Msg.MessageType_MsgType_Order
//			orderChangeReq.Scene = Msg.MessageScene_MsgScene_S2C
//
//			orderProduct := Order.OrderProductBody{}
//			orderProduct.OrderID = newOrder.OrderId
//			orderProduct.State = global.OrderState(newOrder.OrderStatus)
//
//			orderChangeReq.Body, err = proto.Marshal(&orderProduct)
//
//			if err != nil {
//				pc.logger.Error("序列化protobuf order 失败")
//				return
//			}
//
//			//orderChangeReq.Time
//
//			// 系统到商户
//			err1 := pc.nsqClient.BroadcastSystemMsgToAllDevices(orderChangeReq, username)
//			if err1 != nil {
//				pc.logger.Error("推送订单变更事件到商户失败")
//			}
//			err2 := pc.nsqClient.BroadcastSystemMsgToAllDevices(orderChangeReq, req.UserID)
//			if err2 != nil {
//				pc.logger.Error("推送订单变更事件到用户失败")
//
//			}
//		}()
//		RespOk(context, http.StatusOK, codes.SUCCESS)
//		return
//	}
//	RespFail(context, http.StatusOK, codes.InvalidParams, "信息错误")
//	return
//}

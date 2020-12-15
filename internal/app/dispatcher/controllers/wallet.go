/*
这个文件是和钱包相关的restful接口- 钱包及支付宝微信支付模块，/v1/wallet/....
*/
package controllers

import (
	"fmt"
	"net/http"
	// "time"

	Wallet "github.com/lianmi/servers/api/proto/wallet"

	jwt_v2 "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/common/codes"
	"github.com/smartwalle/alipay/v3"
	"go.uber.org/zap"
)

//根据用户注册id获取用户资料
func (pc *LianmiApisController) PreAlipay(c *gin.Context) {
	pc.logger.Debug("PreAlipay start ...")

	claims := jwt_v2.ExtractClaims(c)
	userName := claims[LMCommon.IdentityKey].(string)
	// deviceID := claims["deviceID"].(string)
	// token := jwt_v2.GetToken(c)

	code := codes.InvalidParams
	var req Wallet.PreAlipayReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("PreAlipay, binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {
		if req.TotalAmount == "" {
			RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段: TotalAmount")
		}
		if resp, err := pc.service.PreAlipay(userName, req.TotalAmount); err != nil {
			code = codes.ERROR
			RespFail(c, http.StatusBadRequest, code, "PreAlipay error")
			return
		} else {

			RespData(c, http.StatusBadRequest, http.StatusOK, resp)
		}

	}
}

//支付回调

func (pc *LianmiApisController) AlipayCallback(c *gin.Context) {
	pc.logger.Debug("AlipayCallback start ...")
	var err error
	var aliClient *alipay.Client
	// 第三个参数是沙箱(false) , 正式环境是 true
	if aliClient, err = alipay.New(LMCommon.AlipayAppId, LMCommon.AppPrivateKey, true); err != nil {
		pc.logger.Error("初始化支付宝失败", zap.Error(err))
		return
	}

	//使用支付宝公钥, 只能二选一 , 所以我选了支付宝公钥
	if err = aliClient.LoadAliPayPublicKey(LMCommon.AlipayPublicKey); err != nil {
		pc.logger.Error("加载支付宝公钥发生错误", zap.Error(err))
		return
	} else {
		pc.logger.Debug("加载支付宝公钥成功")
	}

	c.Request.ParseForm()

	var outTradeNo = c.Request.Form.Get("out_trade_no")
	var p = alipay.TradeQuery{}
	p.OutTradeNo = outTradeNo
	rsp, err := aliClient.TradeQuery(p)
	if err != nil {
		errMsg := fmt.Sprintf("AlipayCallback, 验证订单 %s 信息发生错误: %s", outTradeNo, err.Error())
		pc.logger.Error(errMsg)
		return
	}
	if rsp.IsSuccess() == false {
		errMsg := fmt.Sprintf("AlipayCallback, 验证订单 %s 信息发生错误: %s-%s", outTradeNo, rsp.Content.Msg, rsp.Content.SubMsg)
		c.String(http.StatusBadRequest, "AlipayCallback, 验证订单 %s 信息发生错误: %s-%s", outTradeNo, rsp.Content.Msg, rsp.Content.SubMsg)
		pc.logger.Error(errMsg)
		return
	}

	c.String(http.StatusOK, "AlipayCallback, 订单 %s 支付成功", outTradeNo)

	//TODO
	pc.service.AlipayDone(outTradeNo)

}

func (pc *LianmiApisController) AlipayNotify(c *gin.Context) {
	pc.logger.Debug("AlipayCallback start ...")

	var err error
	var aliClient *alipay.Client
	// 第三个参数是沙箱(false) , 正式环境是 true
	if aliClient, err = alipay.New(LMCommon.AlipayAppId, LMCommon.AppPrivateKey, true); err != nil {
		pc.logger.Error("初始化支付宝失败", zap.Error(err))
		return
	}

	//使用支付宝公钥, 只能二选一 , 所以我选了支付宝公钥
	if err = aliClient.LoadAliPayPublicKey(LMCommon.AlipayPublicKey); err != nil {
		pc.logger.Error("加载支付宝公钥发生错误", zap.Error(err))
		return
	} else {
		pc.logger.Debug("加载支付宝公钥成功")
	}

	c.Request.ParseForm()

	ok, err := aliClient.VerifySign(c.Request.Form)
	if err != nil {
		pc.logger.Error("异步通知验证签名发生错误", zap.Error(err))
		return
	}

	if ok == false {
		pc.logger.Error("异步通知验证签名未通过")
		return
	}

	pc.logger.Debug("异步通知验证签名通过")

	var outTradeNo = c.Request.Form.Get("out_trade_no")
	var p = alipay.TradeQuery{}
	p.OutTradeNo = outTradeNo
	rsp, err := aliClient.TradeQuery(p)
	if err != nil {
		errMsg := fmt.Sprintf("AlipayNotify, 异步通知验证订单 %s 信息发生错误: %s ", outTradeNo, err.Error())
		pc.logger.Error(errMsg)
		return
	}
	if rsp.IsSuccess() == false {
		errMsg := fmt.Sprintf("AlipayNotify,异步通知验证订单 %s 信息发生错误: %s-%s \n", outTradeNo, rsp.Content.Msg, rsp.Content.SubMsg)
		pc.logger.Error(errMsg)
		return
	}

	pc.logger.Debug(fmt.Sprintf("AlipayNotify, 订单 %s 支付成功", outTradeNo))

}

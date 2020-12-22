/*
这个文件是和前端相关的restful接口-会员费分销模块，/v1/membership/....
*/
package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	Auth "github.com/lianmi/servers/api/proto/auth"
	// Global "github.com/lianmi/servers/api/proto/global"

	jwt_v2 "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/lianmi/servers/internal/common"
	"go.uber.org/zap"
)

//查询VIP会员价格表
func (pc *LianmiApisController) GetVipPriceList(c *gin.Context) {
	var payType int
	payTypeStr := c.DefaultQuery("pay_type", "0")
	payType, _ = strconv.Atoi(payTypeStr)
	pc.logger.Debug("GetVipPriceList", zap.String("payTypeStr", payTypeStr))

	resp, err := pc.service.GetVipPriceList(payType)

	if err != nil {
		RespFail(c, http.StatusBadRequest, 400, "GetVipPriceList failed")
	} else {

		RespData(c, http.StatusOK, 200, resp)
	}
}

//商户查询当前名下用户总数，按月统计付费会员总数及返佣金额，是否已经返佣
func (pc *LianmiApisController) GetBusinessMembership(c *gin.Context) {
	var req Auth.GetBusinessMembershipReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {

		resp, err := pc.service.GetBusinessMembership(req.BusinessUsername)

		if err != nil {
			RespFail(c, http.StatusBadRequest, 400, "Get BusinessMembership failed")
		} else {

			RespData(c, http.StatusOK, 200, resp)
		}
	}

}

//预生成一个购买会员的订单， 返回OrderID及预转账裸交易数据
func (pc *LianmiApisController) PreOrderForPayMembership(c *gin.Context) {
	claims := jwt_v2.ExtractClaims(c)
	userName := claims[common.IdentityKey].(string)
	deviceID := claims["deviceID"].(string)
	token := jwt_v2.GetToken(c)

	var req Auth.PreOrderForPayMembershipReq

	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {
		//如果目标用户是空，这为自己购买
		if req.PayForUsername == "" {
			req.PayForUsername = userName
		}
		if req.PayType == 0 {
			RespFail(c, http.StatusBadRequest, 500, "PayType is zero")
		}
		pc.logger.Debug("PreOrderForPayMembership",
			zap.String("userName", userName),
			zap.String("deviceID", deviceID),
			zap.String("payForUsername", req.PayForUsername), //要给谁付费, 如果是给自己，则留空或填自己的注册账号
			zap.Int("PayType", int(req.PayType)),             //枚举 购买的会员类型，月卡、 季卡或年卡
			zap.String("token", token))

		ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
		resp, err := pc.service.PreOrderForPayMembership(ctx, userName, deviceID, &req)

		if err != nil {
			RespFail(c, http.StatusBadRequest, 400, "PreOrderForPayMembership failed")
		} else {
			RespData(c, http.StatusOK, 200, resp)
		}
	}

}

//确认购买会员
//调用此接口前，需要调用 PreOrderForPayMembership 发起会员付费转账,
//在本地签名，然后携带签名后的交易数据提交到服务端，返回区块高度，交易哈希
//会员付费， 可以他人代付， 如果他人代付，自动成为其推荐人, 强制归属同一个商户,
//支付成功后，向用户发出通知
//如果用户是自行注册的，提醒用户输入商户的推荐码
func (pc *LianmiApisController) ConfirmPayForMembership(c *gin.Context) {
	claims := jwt_v2.ExtractClaims(c)
	userName := claims[common.IdentityKey].(string)
	deviceID := claims["deviceID"].(string)
	token := jwt_v2.GetToken(c)

	pc.logger.Debug("ConfirmPayForMembership",
		zap.String("userName", userName),
		zap.String("deviceID", deviceID),
		zap.String("token", token))

	var req Auth.ConfirmPayForMembershipReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {
		ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
		resp, err := pc.service.ConfirmPayForMembership(ctx, userName, &req)

		if err != nil {
			RespFail(c, http.StatusBadRequest, 400, "PayForMembership failed")
		} else {
			RespData(c, http.StatusOK, 200, resp)
		}
	}
}

func (pc *LianmiApisController) GetNormalMembership(c *gin.Context) {

	var req Auth.GetMembershipReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {

		resp, err := pc.service.GetNormalMembership(req.Username)

		if err != nil {
			RespFail(c, http.StatusBadRequest, 400, "Get Membership failed")
		} else {

			RespData(c, http.StatusOK, 200, resp)
		}
	}
}

//提交佣金提现申请(商户，用户)
func (pc *LianmiApisController) SubmitCommssionWithdraw(c *gin.Context) {

	var req Auth.CommssionWithdrawReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {
		if req.Username == "" || req.YearMonth == "" {
			RespFail(c, http.StatusBadRequest, 400, "Submit Commssion Withdraw failed")
		}

		resp, err := pc.service.SubmitCommssionWithdraw(req.Username, req.YearMonth)

		if err != nil {
			RespFail(c, http.StatusBadRequest, 400, "Submit Commssion Withdraw failed")
		} else {

			RespData(c, http.StatusOK, 200, resp)
		}
	}

}

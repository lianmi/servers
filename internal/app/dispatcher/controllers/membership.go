/*
这个文件是和前端相关的restful接口-会员费分销模块，/v1/membership/....
*/
package controllers

import (
	"net/http"

	Global "github.com/lianmi/servers/api/proto/global"
	// Order "github.com/lianmi/servers/api/proto/order"
	Auth "github.com/lianmi/servers/api/proto/auth"

	jwt_v2 "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/lianmi/servers/internal/common"
	"go.uber.org/zap"
)

func (pc *LianmiApisController) GetMembershipCardSaleMode(c *gin.Context) {
	var req Auth.MembershipCardSaleModeReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {

		if req.BusinessUsername == "" {
			RespFail(c, http.StatusBadRequest, 400, "BusinessUsername参数错误")
		}
		saleMode, err := pc.service.GetMembershipCardSaleMode(req.BusinessUsername)

		if err != nil {
			RespFail(c, http.StatusBadRequest, 400, "Get Membership Card Sale Mode failed")
		} else {

			RespData(c, http.StatusOK, 200, &Auth.MembershipCardSaleModeResp{
				SaleType: Global.MembershipCardSaleType(saleMode),
			})
		}
	}

}

func (pc *LianmiApisController) SetMembershipCardSaleMode(c *gin.Context) {
	var req Auth.MembershipCardSaleModeReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {

		if req.BusinessUsername == "" {
			RespFail(c, http.StatusBadRequest, 400, "BusinessUsername参数错误")
		}

		saleType := int(req.SaleType)
		if saleType == 0 {
			saleType = 1
		}

		if !(saleType == 1 || saleType == 2) {
			RespFail(c, http.StatusBadRequest, 400, "Set Membership Card Sale Mode failed")
		}

		err := pc.service.SetMembershipCardSaleMode(req.BusinessUsername, saleType)

		if err != nil {
			RespFail(c, http.StatusBadRequest, 400, "Set Membership Card Sale Mode failed")
		} else {

			RespOk(c, http.StatusOK, 200)
		}
	}

}

func (pc *LianmiApisController) GetBusinessMembership(c *gin.Context) {
	var req Auth.GetBusinessMembershipReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {

		resp, err := pc.service.GetBusinessMembership(req.IsRebate)

		if err != nil {
			RespFail(c, http.StatusBadRequest, 400, "Get BusinessMembership failed")
		} else {
			// resp := &Auth.GetBusinessMembershipResp{
			// 	Totalmembers: 1,
			// }
			//TODO
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

	pc.logger.Debug("PreOrderForPayMembership",
		zap.String("userName", userName),
		zap.String("deviceID", deviceID),
		zap.String("token", token))

	err := pc.service.PreOrderForPayMembership(userName, deviceID)

	if err != nil {
		RespFail(c, http.StatusBadRequest, 400, "PreOrderForPayMembership failed")
	} else {
		RespData(c, http.StatusOK, 200, &Auth.PreOrderForPayMembershipResp{
			//TODO
		})
	}

}

//调用此接口前，需要调用 PreOrderForPayMembership 发起会员付费转账,
//在本地签名，然后携带签名后的交易数据提交到服务端，返回区块高度，交易哈希
//会员付费， 可以他人代付， 如果他人代付，自动成为其推荐人, 强制归属同一个商户,
//支付成功后，向用户发出通知
//如果用户是自行注册的，提醒用户输入商户的推荐码
func (pc *LianmiApisController) PayForMembership(c *gin.Context) {
	claims := jwt_v2.ExtractClaims(c)
	userName := claims[common.IdentityKey].(string)
	deviceID := claims["deviceID"].(string)
	token := jwt_v2.GetToken(c)

	pc.logger.Debug("PayForMembership",
		zap.String("userName", userName),
		zap.String("deviceID", deviceID),
		zap.String("token", token))

	var req Auth.PayForMembershipReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {

		err := pc.service.PayForMembership(req.PayUsername)

		if err != nil {
			RespFail(c, http.StatusBadRequest, 400, "PayForMembership failed")
		} else {
			RespOk(c, http.StatusOK, 200)
		}
	}
}

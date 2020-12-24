/*
这个文件是和前端相关的restful接口-会员费分销模块，/v1/membership/....
*/
package controllers

import (
	// "context"
	"net/http"
	"strconv"
	// "time"

	Auth "github.com/lianmi/servers/api/proto/auth"
	// Global "github.com/lianmi/servers/api/proto/global"

	// jwt_v2 "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	// "github.com/lianmi/servers/internal/common"
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

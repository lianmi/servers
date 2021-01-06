package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/global"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/model/request"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/model/response"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/service"
	"github.com/lianmi/servers/internal/pkg/models"
	"go.uber.org/zap"
)

// @Tags LotterySaleTimes
// @Summary 创建LotterySaleTimes
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body models.LotterySaleTime true "创建LotterySaleTimes"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /lotterySaleTimes/createLotterySaleTimes [post]
func CreateLotterySaleTimes(c *gin.Context) {
	var lotterySaleTimes models.LotterySaleTime
	_ = c.ShouldBindJSON(&lotterySaleTimes)
	if err := service.CreateLotterySaleTimes(lotterySaleTimes); err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Any("err", err))
		response.FailWithMessage("创建失败", c)
	} else {
		response.OkWithMessage("创建成功", c)
	}
}

// @Tags LotterySaleTimes
// @Summary 删除LotterySaleTimes
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body models.LotterySaleTime true "删除LotterySaleTimes"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /lotterySaleTimes/deleteLotterySaleTimes [delete]
func DeleteLotterySaleTimes(c *gin.Context) {
	var lotterySaleTimes models.LotterySaleTime
	_ = c.ShouldBindJSON(&lotterySaleTimes)
	if err := service.DeleteLotterySaleTimes(lotterySaleTimes); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Any("err", err))
		response.FailWithMessage("删除失败", c)
	} else {
		response.OkWithMessage("删除成功", c)
	}
}

// @Tags LotterySaleTimes
// @Summary 批量删除LotterySaleTimes
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除LotterySaleTimes"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"批量删除成功"}"
// @Router /lotterySaleTimes/deleteLotterySaleTimesByIds [delete]
func DeleteLotterySaleTimesByIds(c *gin.Context) {
	var IDS request.IdsReq
	_ = c.ShouldBindJSON(&IDS)
	if err := service.DeleteLotterySaleTimesByIds(IDS); err != nil {
		global.GVA_LOG.Error("批量删除失败!", zap.Any("err", err))
		response.FailWithMessage("批量删除失败", c)
	} else {
		response.OkWithMessage("批量删除成功", c)
	}
}

// @Tags LotterySaleTimes
// @Summary 更新LotterySaleTimes
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body models.LotterySaleTime true "更新LotterySaleTimes"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /lotterySaleTimes/updateLotterySaleTimes [put]
func UpdateLotterySaleTimes(c *gin.Context) {
	var lotterySaleTimes models.LotterySaleTime
	_ = c.ShouldBindJSON(&lotterySaleTimes)
	if err := service.UpdateLotterySaleTimes(lotterySaleTimes); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Any("err", err))
		response.FailWithMessage("更新失败", c)
	} else {
		response.OkWithMessage("更新成功", c)
	}
}

// @Tags LotterySaleTimes
// @Summary 用id查询LotterySaleTimes
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body models.LotterySaleTime true "用id查询LotterySaleTimes"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /lotterySaleTimes/findLotterySaleTimes [get]
func FindLotterySaleTimes(c *gin.Context) {
	var lotterySaleTimes models.LotterySaleTime
	_ = c.ShouldBindQuery(&lotterySaleTimes)
	if err, relotterySaleTimes := service.GetLotterySaleTimes(lotterySaleTimes.ID); err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Any("err", err))
		response.FailWithMessage("查询失败", c)
	} else {
		response.OkWithData(gin.H{"relotterySaleTimes": relotterySaleTimes}, c)
	}
}

// @Tags LotterySaleTimes
// @Summary 分页获取LotterySaleTimes列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body request.LotterySaleTimesSearch true "分页获取LotterySaleTimes列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /lotterySaleTimes/getLotterySaleTimesList [get]
func GetLotterySaleTimesList(c *gin.Context) {
	var pageInfo request.LotterySaleTimesSearch
	_ = c.ShouldBindQuery(&pageInfo)
	if err, list, total := service.GetLotterySaleTimesInfoList(pageInfo); err != nil {
		global.GVA_LOG.Error("获取失败", zap.Any("err", err))
		response.FailWithMessage("获取失败", c)
	} else {
		response.OkWithDetailed(response.PageResult{
			List:     list,
			Total:    total,
			Page:     pageInfo.Page,
			PageSize: pageInfo.PageSize,
		}, "获取成功", c)
	}
}

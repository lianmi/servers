package router

import (
	"github.com/lianmi/servers/internal/app/gin-vue-admin/api/v1"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/middleware"
	"github.com/gin-gonic/gin"
)

func InitLotterySaleTimesRouter(Router *gin.RouterGroup) {
	LotterySaleTimesRouter := Router.Group("lotterySaleTimes").Use(middleware.JWTAuth()).Use(middleware.CasbinHandler()).Use(middleware.OperationRecord())
	{
		LotterySaleTimesRouter.POST("createLotterySaleTimes", v1.CreateLotterySaleTimes)   // 新建LotterySaleTimes
		LotterySaleTimesRouter.DELETE("deleteLotterySaleTimes", v1.DeleteLotterySaleTimes) // 删除LotterySaleTimes
		LotterySaleTimesRouter.DELETE("deleteLotterySaleTimesByIds", v1.DeleteLotterySaleTimesByIds) // 批量删除LotterySaleTimes
		LotterySaleTimesRouter.PUT("updateLotterySaleTimes", v1.UpdateLotterySaleTimes)    // 更新LotterySaleTimes
		LotterySaleTimesRouter.GET("findLotterySaleTimes", v1.FindLotterySaleTimes)        // 根据ID获取LotterySaleTimes
		LotterySaleTimesRouter.GET("getLotterySaleTimesList", v1.GetLotterySaleTimesList)  // 获取LotterySaleTimes列表
	}
}

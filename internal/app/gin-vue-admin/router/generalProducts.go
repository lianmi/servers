package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/api/v1"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/middleware"
)

func InitGeneralProductRouter(Router *gin.RouterGroup) {
	GeneralProductRouter := Router.Group("generalProducts").Use(middleware.JWTAuth()).Use(middleware.CasbinHandler()).Use(middleware.OperationRecord())
	{
		GeneralProductRouter.POST("createGeneralProduct", v1.CreateGeneralProduct)             // 新建GeneralProduct
		GeneralProductRouter.DELETE("deleteGeneralProduct", v1.DeleteGeneralProduct)           // 删除GeneralProduct
		GeneralProductRouter.DELETE("deleteGeneralProductByIds", v1.DeleteGeneralProductByIds) // 批量删除GeneralProduct
		GeneralProductRouter.PUT("updateGeneralProduct", v1.UpdateGeneralProduct)              // 更新GeneralProduct
		GeneralProductRouter.GET("findGeneralProduct", v1.FindGeneralProduct)                  // 根据ID获取GeneralProduct
		GeneralProductRouter.GET("getGeneralProductList", v1.GetGeneralProductList)            // 获取GeneralProduct列表
	}
}

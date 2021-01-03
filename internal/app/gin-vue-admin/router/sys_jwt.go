package router

import (
	"github.com/lianmi/servers/internal/app/gin-vue-admin/api/v1"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/middleware"
	"github.com/gin-gonic/gin"
)

func InitJwtRouter(Router *gin.RouterGroup) {
	ApiRouter := Router.Group("jwt").Use(middleware.OperationRecord())
	{
		ApiRouter.POST("jsonInBlacklist", v1.JsonInBlacklist) // jwt加入黑名单
	}
}

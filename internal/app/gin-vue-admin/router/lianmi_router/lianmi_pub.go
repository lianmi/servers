package lianmi_router

import (
	"github.com/gin-gonic/gin"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/api/v1/lianmiApi"
)

func InitLianmiPubRouter(Router *gin.RouterGroup) (R gin.IRoutes) {
	PubRouter := Router.Group("")
	{
		PubRouter.GET("about", lianmiApi.About)
	}
	return PubRouter
}

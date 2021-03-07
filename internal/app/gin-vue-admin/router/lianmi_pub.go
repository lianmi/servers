package router

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/lianmi/servers/internal/app/gin-vue-admin/api/v1"
)

func InitLianmiPubRouter(Router *gin.RouterGroup) (R gin.IRoutes) {
	PubRouter := Router.Group("")
	{
		PubRouter.GET("about", v1.About)
	}
	return PubRouter
}

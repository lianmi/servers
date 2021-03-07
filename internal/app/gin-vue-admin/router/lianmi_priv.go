package router

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/lianmi/servers/internal/app/gin-vue-admin/api/v1"
)

func InitLianmiPrivRouter(Router *gin.RouterGroup) (R gin.IRoutes) {
	LianmiPrivRouter := Router.Group("")
	{
		LianmiPrivRouter.GET("about", v1.About)
	}
	return LianmiPrivRouter
}

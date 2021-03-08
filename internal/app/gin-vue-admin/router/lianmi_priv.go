/*

此文件用于lianmi接口的路由

*/
package router

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/lianmi/servers/internal/app/gin-vue-admin/api/v1"
)

func InitLianmiPrivRouter(Router *gin.RouterGroup) (R gin.IRoutes) {
	LianmiPrivRouter := Router.Group("lianmi")
	{
		//用户模块
		LianmiPrivRouter.POST("getUsers", v1.LianmiGetUsers)

	}
	return LianmiPrivRouter
}

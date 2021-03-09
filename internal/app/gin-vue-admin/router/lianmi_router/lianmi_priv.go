/*

此文件用于lianmi接口的路由

*/
package lianmi_router

import (
	"github.com/gin-gonic/gin"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/api/v1/lianmiApi"
)

func InitLianmiPrivRouter(Router *gin.RouterGroup) (R gin.IRoutes) {
	LianmiPrivRouter := Router.Group("lianmi")
	{
		//用户模块
		LianmiPrivRouter.POST("getUsers", lianmiApi.LianmiGetUsers)

	}
	return LianmiPrivRouter
}

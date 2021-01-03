package router

import (
	"github.com/lianmi/servers/internal/app/gin-vue-admin/api/v1"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/middleware"
	"github.com/gin-gonic/gin"
)

func InitEmailRouter(Router *gin.RouterGroup) {
	UserRouter := Router.Group("email").Use(middleware.OperationRecord())
	{
		UserRouter.POST("emailTest", v1.EmailTest) // 发送测试邮件
	}
}

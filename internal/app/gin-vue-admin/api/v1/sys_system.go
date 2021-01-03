package v1

import (
	"github.com/lianmi/servers/internal/app/gin-vue-admin/global"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/model"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/model/response"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Tags System
// @Summary 获取配置文件内容
// @Security ApiKeyAuth
// @Produce  application/json
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /system/getSystemConfig [post]
func GetSystemConfig(c *gin.Context) {
	if err, config := service.GetSystemConfig(); err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Any("err", err))
		response.FailWithMessage("获取失败", c)
	} else {
		response.OkWithDetailed(response.SysConfigResponse{Config: config}, "获取成功", c)
	}
}

// @Tags System
// @Summary 设置配置文件内容
// @Security ApiKeyAuth
// @Produce  application/json
// @Param data body model.System true "设置配置文件内容"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"设置成功"}"
// @Router /system/setSystemConfig [post]
func SetSystemConfig(c *gin.Context) {
	var sys model.System
	_ = c.ShouldBindJSON(&sys)
	if err := service.SetSystemConfig(sys); err != nil {
		global.GVA_LOG.Error("设置失败!", zap.Any("err", err))
		response.FailWithMessage("设置失败", c)
	} else {
		response.OkWithData("设置成功", c)
	}
}

// 本方法开发中 开发者windows系统 缺少linux系统所需的包 因此搁置
// @Tags System
// @Summary 重启系统
// @Security ApiKeyAuth
// @Produce  application/json
// @Param data body model.System true "重启系统"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"重启系统成功"}"
// @Router /system/ReloadSystem [post]
func ReloadSystem(c *gin.Context) {
	var sys model.System
	_ = c.ShouldBindJSON(&sys)
	if err := service.SetSystemConfig(sys); err != nil {
		global.GVA_LOG.Error("重启系统失败!", zap.Any("err", err))
		response.FailWithMessage("重启系统失败", c)
	} else {
		response.OkWithMessage("重启系统成功", c)
	}
}

// @Tags System
// @Summary 获取服务器信息
// @Security ApiKeyAuth
// @Produce  application/json
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /system/getServerInfo [post]
func GetServerInfo(c *gin.Context) {
	if server, err := service.GetServerInfo(); err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Any("err", err))
		response.FailWithMessage("获取失败", c)
		return
	} else {
		response.OkWithDetailed(gin.H{"server": server}, "获取成功", c)
	}

}

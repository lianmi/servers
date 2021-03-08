/*
用户模块
CRUD
*/

package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/global"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/model/request"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/model/response"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/service"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/utils"
	"go.uber.org/zap"
)

// @Tags LianmiUser
// @Summary 分页获取用户列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body request.PageInfo true "页码, 每页大小"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /lianmi/getUsers [post]
func LianmiGetUsers(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindJSON(&pageInfo)
	if err := utils.Verify(pageInfo, utils.PageInfoVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err, list, total := service.LianmiGetUsers(pageInfo); err != nil {
		global.GVA_LOG.Error("获取失败", zap.Any("err", err))
		response.FailWithMessage("获取失败", c)
	} else {
		response.OkWithDetailed(response.PageResult{
			List:     list,
			Total:    total,
			Page:     pageInfo.Page,
			PageSize: pageInfo.PageSize,
		}, "获取成功", c)
	}
}

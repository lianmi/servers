/*
公有接口，关于
*/
package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/model/response"
)

// @Tags About
// @Summary 关于
// @Produce  application/json
// @Success 200 {string} string "{"code":0,"data":{},"msg":"lianmi dashboard v0.1"}"
// @Router /about [get]
func About(c *gin.Context) {
	response.OkWithMessage("lianmi dashboard v0.1", c)
}

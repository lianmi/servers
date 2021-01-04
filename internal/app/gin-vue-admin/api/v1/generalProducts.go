package v1

import (
	"github.com/lianmi/servers/internal/app/gin-vue-admin/global"
    "github.com/lianmi/servers/internal/app/gin-vue-admin/model"
    "github.com/lianmi/servers/internal/app/gin-vue-admin/model/request"
    "github.com/lianmi/servers/internal/app/gin-vue-admin/model/response"
    "github.com/lianmi/servers/internal/app/gin-vue-admin/service"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// @Tags GeneralProduct
// @Summary 创建GeneralProduct
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.GeneralProduct true "创建GeneralProduct"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /generalProducts/createGeneralProduct [post]
func CreateGeneralProduct(c *gin.Context) {
	var generalProducts model.GeneralProduct
	_ = c.ShouldBindJSON(&generalProducts)
	if err := service.CreateGeneralProduct(generalProducts); err != nil {
        global.GVA_LOG.Error("创建失败!", zap.Any("err", err))
		response.FailWithMessage("创建失败", c)
	} else {
		response.OkWithMessage("创建成功", c)
	}
}

// @Tags GeneralProduct
// @Summary 删除GeneralProduct
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.GeneralProduct true "删除GeneralProduct"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /generalProducts/deleteGeneralProduct [delete]
func DeleteGeneralProduct(c *gin.Context) {
	var generalProducts model.GeneralProduct
	_ = c.ShouldBindJSON(&generalProducts)
	if err := service.DeleteGeneralProduct(generalProducts); err != nil {
        global.GVA_LOG.Error("删除失败!", zap.Any("err", err))
		response.FailWithMessage("删除失败", c)
	} else {
		response.OkWithMessage("删除成功", c)
	}
}

// @Tags GeneralProduct
// @Summary 批量删除GeneralProduct
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除GeneralProduct"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"批量删除成功"}"
// @Router /generalProducts/deleteGeneralProductByIds [delete]
func DeleteGeneralProductByIds(c *gin.Context) {
	var IDS request.IdsReq
    _ = c.ShouldBindJSON(&IDS)
	if err := service.DeleteGeneralProductByIds(IDS); err != nil {
        global.GVA_LOG.Error("批量删除失败!", zap.Any("err", err))
		response.FailWithMessage("批量删除失败", c)
	} else {
		response.OkWithMessage("批量删除成功", c)
	}
}

// @Tags GeneralProduct
// @Summary 更新GeneralProduct
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.GeneralProduct true "更新GeneralProduct"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /generalProducts/updateGeneralProduct [put]
func UpdateGeneralProduct(c *gin.Context) {
	var generalProducts model.GeneralProduct
	_ = c.ShouldBindJSON(&generalProducts)
	if err := service.UpdateGeneralProduct(generalProducts); err != nil {
        global.GVA_LOG.Error("更新失败!", zap.Any("err", err))
		response.FailWithMessage("更新失败", c)
	} else {
		response.OkWithMessage("更新成功", c)
	}
}

// @Tags GeneralProduct
// @Summary 用id查询GeneralProduct
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.GeneralProduct true "用id查询GeneralProduct"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /generalProducts/findGeneralProduct [get]
func FindGeneralProduct(c *gin.Context) {
	var generalProducts model.GeneralProduct
	_ = c.ShouldBindQuery(&generalProducts)
	if err, regeneralProducts := service.GetGeneralProduct(generalProducts.ID); err != nil {
        global.GVA_LOG.Error("查询失败!", zap.Any("err", err))
		response.FailWithMessage("查询失败", c)
	} else {
		response.OkWithData(gin.H{"regeneralProducts": regeneralProducts}, c)
	}
}

// @Tags GeneralProduct
// @Summary 分页获取GeneralProduct列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body request.GeneralProductSearch true "分页获取GeneralProduct列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /generalProducts/getGeneralProductList [get]
func GetGeneralProductList(c *gin.Context) {
	var pageInfo request.GeneralProductSearch
	_ = c.ShouldBindQuery(&pageInfo)
	if err, list, total := service.GetGeneralProductInfoList(pageInfo); err != nil {
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

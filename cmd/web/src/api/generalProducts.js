import service from '@/utils/request'

// @Tags GeneralProduct
// @Summary 创建GeneralProduct
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.GeneralProduct true "创建GeneralProduct"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /generalProducts/createGeneralProduct [post]
export const createGeneralProduct = (data) => {
     return service({
         url: "/generalProducts/createGeneralProduct",
         method: 'post',
         data
     })
 }


// @Tags GeneralProduct
// @Summary 删除GeneralProduct
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.GeneralProduct true "删除GeneralProduct"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /generalProducts/deleteGeneralProduct [delete]
 export const deleteGeneralProduct = (data) => {
     return service({
         url: "/generalProducts/deleteGeneralProduct",
         method: 'delete',
         data
     })
 }

// @Tags GeneralProduct
// @Summary 删除GeneralProduct
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除GeneralProduct"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /generalProducts/deleteGeneralProduct [delete]
 export const deleteGeneralProductByIds = (data) => {
     return service({
         url: "/generalProducts/deleteGeneralProductByIds",
         method: 'delete',
         data
     })
 }

// @Tags GeneralProduct
// @Summary 更新GeneralProduct
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.GeneralProduct true "更新GeneralProduct"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /generalProducts/updateGeneralProduct [put]
 export const updateGeneralProduct = (data) => {
     return service({
         url: "/generalProducts/updateGeneralProduct",
         method: 'put',
         data
     })
 }


// @Tags GeneralProduct
// @Summary 用id查询GeneralProduct
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.GeneralProduct true "用id查询GeneralProduct"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /generalProducts/findGeneralProduct [get]
 export const findGeneralProduct = (params) => {
     return service({
         url: "/generalProducts/findGeneralProduct",
         method: 'get',
         params
     })
 }


// @Tags GeneralProduct
// @Summary 分页获取GeneralProduct列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body request.PageInfo true "分页获取GeneralProduct列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /generalProducts/getGeneralProductList [get]
 export const getGeneralProductList = (params) => {
     return service({
         url: "/generalProducts/getGeneralProductList",
         method: 'get',
         params
     })
 }
import service from '@/utils/request'

// @Tags LotterySaleTimes
// @Summary 创建LotterySaleTimes
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.LotterySaleTimes true "创建LotterySaleTimes"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /lotterySaleTimes/createLotterySaleTimes [post]
export const createLotterySaleTimes = (data) => {
     return service({
         url: "/lotterySaleTimes/createLotterySaleTimes",
         method: 'post',
         data
     })
 }


// @Tags LotterySaleTimes
// @Summary 删除LotterySaleTimes
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.LotterySaleTimes true "删除LotterySaleTimes"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /lotterySaleTimes/deleteLotterySaleTimes [delete]
 export const deleteLotterySaleTimes = (data) => {
     return service({
         url: "/lotterySaleTimes/deleteLotterySaleTimes",
         method: 'delete',
         data
     })
 }

// @Tags LotterySaleTimes
// @Summary 删除LotterySaleTimes
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除LotterySaleTimes"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /lotterySaleTimes/deleteLotterySaleTimes [delete]
 export const deleteLotterySaleTimesByIds = (data) => {
     return service({
         url: "/lotterySaleTimes/deleteLotterySaleTimesByIds",
         method: 'delete',
         data
     })
 }

// @Tags LotterySaleTimes
// @Summary 更新LotterySaleTimes
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.LotterySaleTimes true "更新LotterySaleTimes"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /lotterySaleTimes/updateLotterySaleTimes [put]
 export const updateLotterySaleTimes = (data) => {
     return service({
         url: "/lotterySaleTimes/updateLotterySaleTimes",
         method: 'put',
         data
     })
 }


// @Tags LotterySaleTimes
// @Summary 用id查询LotterySaleTimes
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.LotterySaleTimes true "用id查询LotterySaleTimes"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /lotterySaleTimes/findLotterySaleTimes [get]
 export const findLotterySaleTimes = (params) => {
     return service({
         url: "/lotterySaleTimes/findLotterySaleTimes",
         method: 'get',
         params
     })
 }


// @Tags LotterySaleTimes
// @Summary 分页获取LotterySaleTimes列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body request.PageInfo true "分页获取LotterySaleTimes列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /lotterySaleTimes/getLotterySaleTimesList [get]
 export const getLotterySaleTimesList = (params) => {
     return service({
         url: "/lotterySaleTimes/getLotterySaleTimesList",
         method: 'get',
         params
     })
 }
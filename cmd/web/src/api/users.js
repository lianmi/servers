import service from '@/utils/request'

// @Tags 用户表 
// @Summary 创建用户表 
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.用户表  true "创建用户表 "
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /users/create用户表  [post]
export const create用户表  = (data) => {
     return service({
         url: "/users/create用户表 ",
         method: 'post',
         data
     })
 }


// @Tags 用户表 
// @Summary 删除用户表 
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.用户表  true "删除用户表 "
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /users/delete用户表  [delete]
 export const delete用户表  = (data) => {
     return service({
         url: "/users/delete用户表 ",
         method: 'delete',
         data
     })
 }

// @Tags 用户表 
// @Summary 删除用户表 
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除用户表 "
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /users/delete用户表  [delete]
 export const delete用户表 ByIds = (data) => {
     return service({
         url: "/users/delete用户表 ByIds",
         method: 'delete',
         data
     })
 }

// @Tags 用户表 
// @Summary 更新用户表 
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.用户表  true "更新用户表 "
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /users/update用户表  [put]
 export const update用户表  = (data) => {
     return service({
         url: "/users/update用户表 ",
         method: 'put',
         data
     })
 }


// @Tags 用户表 
// @Summary 用id查询用户表 
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.用户表  true "用id查询用户表 "
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /users/find用户表  [get]
 export const find用户表  = (params) => {
     return service({
         url: "/users/find用户表 ",
         method: 'get',
         params
     })
 }


// @Tags 用户表 
// @Summary 分页获取用户表 列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body request.PageInfo true "分页获取用户表 列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /users/get用户表 List [get]
 export const get用户表 List = (params) => {
     return service({
         url: "/users/get用户表 List",
         method: 'get',
         params
     })
 }
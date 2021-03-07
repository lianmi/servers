package v1

// @Tags LianmiUsers
// @Summary 分页获取app用户列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body request.PageInfo true "页码, 每页大小"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /admin/user/getAllUsers [post]
func LianmiGetAllUsers() {

}

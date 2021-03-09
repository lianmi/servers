package lianmiservice

import (
	"fmt"

	"github.com/lianmi/servers/internal/app/gin-vue-admin/global"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/model/request"
	"github.com/lianmi/servers/internal/pkg/models"
)

//@author: lianmi.cloud
//@function: LianmiGetUsers
//@description: 分页获取数据
//@param: info request.PageInfo
//@return: err error, list interface{}, total int64

func LianmiGetUsers(info request.PageInfo) (err error, list interface{}, total int64) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	db := global.LIANMI_DB.Model(&models.User{})
	var userList []models.User
	err = db.Count(&total).Error
	err = db.Limit(limit).Offset(offset).Find(&userList).Error
	global.GVA_LOG.Info(fmt.Sprintf("%d", total))
	return err, userList, total
}

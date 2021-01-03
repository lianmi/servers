package request

import "github.com/lianmi/servers/cmd/gin-vue-admin/model"

// Add menu authority info structure
type AddMenuAuthorityInfo struct {
	Menus       []model.SysBaseMenu
	AuthorityId string
}

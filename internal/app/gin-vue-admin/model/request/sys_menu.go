package request

import "github.com/lianmi/servers/internal/app/gin-vue-admin/model"

// Add menu authority info structure
type AddMenuAuthorityInfo struct {
	Menus       []model.SysBaseMenu
	AuthorityId string
}

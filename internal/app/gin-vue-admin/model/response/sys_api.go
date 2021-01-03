package response

import "github.com/lianmi/servers/internal/app/gin-vue-admin/model"

type SysAPIResponse struct {
	Api model.SysApi `json:"api"`
}

type SysAPIListResponse struct {
	Apis []model.SysApi `json:"apis"`
}

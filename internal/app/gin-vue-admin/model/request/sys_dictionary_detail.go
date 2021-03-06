package request

import "github.com/lianmi/servers/internal/app/gin-vue-admin/model"

type SysDictionaryDetailSearch struct {
	model.SysDictionaryDetail
	PageInfo
}

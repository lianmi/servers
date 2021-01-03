package request

import "github.com/lianmi/servers/cmd/gin-vue-admin/model"

type SysDictionaryDetailSearch struct{
    model.SysDictionaryDetail
    PageInfo
}
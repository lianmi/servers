package request

import "github.com/lianmi/servers/internal/app/gin-vue-admin/model"

type SysDictionarySearch struct{
    model.SysDictionary
    PageInfo
}
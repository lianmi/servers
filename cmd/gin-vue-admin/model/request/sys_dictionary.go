package request

import "github.com/lianmi/servers/cmd/gin-vue-admin/model"

type SysDictionarySearch struct{
    model.SysDictionary
    PageInfo
}
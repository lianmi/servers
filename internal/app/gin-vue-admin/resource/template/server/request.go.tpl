package request

import "github.com/lianmi/servers/cmd/gin-vue-admin/model"

type {{.StructName}}Search struct{
    model.{{.StructName}}
    PageInfo
}
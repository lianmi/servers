package request

import "github.com/lianmi/servers/internal/app/gin-vue-admin/model"

type {{.StructName}}Search struct{
    model.{{.StructName}}
    PageInfo
}
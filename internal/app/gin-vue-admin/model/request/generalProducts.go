package request

import "github.com/lianmi/servers/internal/app/gin-vue-admin/model"

type GeneralProductSearch struct{
    model.GeneralProduct
    PageInfo
}
package request

import "github.com/lianmi/servers/internal/pkg/models"

type GeneralProductSearch struct {
	models.GeneralProduct
	PageInfo
}

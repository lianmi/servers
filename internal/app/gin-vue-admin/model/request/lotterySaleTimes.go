package request

import "github.com/lianmi/servers/internal/pkg/models"

type LotterySaleTimesSearch struct {
	models.LotterySaleTime
	PageInfo
}

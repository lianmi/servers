package response

import "github.com/lianmi/servers/internal/app/gin-vue-admin/model"

type ExaCustomerResponse struct {
	Customer model.ExaCustomer `json:"customer"`
}

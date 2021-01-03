package response

import "github.com/lianmi/servers/cmd/gin-vue-admin/model"

type ExaCustomerResponse struct {
	Customer model.ExaCustomer `json:"customer"`
}

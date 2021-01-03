package response

import "github.com/lianmi/servers/cmd/gin-vue-admin/model/request"

type PolicyPathResponse struct {
	Paths []request.CasbinInfo `json:"paths"`
}

package response

import "github.com/lianmi/servers/internal/app/gin-vue-admin/model/request"

type PolicyPathResponse struct {
	Paths []request.CasbinInfo `json:"paths"`
}

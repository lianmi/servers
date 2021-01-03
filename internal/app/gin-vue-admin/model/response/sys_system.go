package response

import "github.com/lianmi/servers/cmd/gin-vue-admin/config"

type SysConfigResponse struct {
	Config config.Server `json:"config"`
}

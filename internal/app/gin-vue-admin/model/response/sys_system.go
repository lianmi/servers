package response

import "github.com/lianmi/servers/internal/app/gin-vue-admin/config"

type SysConfigResponse struct {
	Config config.Server `json:"config"`
}

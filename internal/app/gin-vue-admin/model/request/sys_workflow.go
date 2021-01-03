package request

import "github.com/lianmi/servers/internal/app/gin-vue-admin/model"

type WorkflowProcessSearch struct {
	model.WorkflowProcess
	PageInfo
}

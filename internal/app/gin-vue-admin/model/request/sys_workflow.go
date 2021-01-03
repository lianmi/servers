package request

import "github.com/lianmi/servers/cmd/gin-vue-admin/model"

type WorkflowProcessSearch struct {
	model.WorkflowProcess
	PageInfo
}

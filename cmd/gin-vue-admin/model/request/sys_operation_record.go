package request

import "github.com/lianmi/servers/cmd/gin-vue-admin/model"

type SysOperationRecordSearch struct {
	model.SysOperationRecord
	PageInfo
}

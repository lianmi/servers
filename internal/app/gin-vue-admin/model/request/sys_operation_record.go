package request

import "github.com/lianmi/servers/internal/app/gin-vue-admin/model"

type SysOperationRecordSearch struct {
	model.SysOperationRecord
	PageInfo
}

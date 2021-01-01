package response

import "github.com/lianmi/servers/cmd/gin-vue-admin/model"

type ExaFileResponse struct {
	File model.ExaFileUploadAndDownload `json:"file"`
}

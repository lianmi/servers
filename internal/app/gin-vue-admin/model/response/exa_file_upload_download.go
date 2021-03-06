package response

import "github.com/lianmi/servers/internal/app/gin-vue-admin/model"

type ExaFileResponse struct {
	File model.ExaFileUploadAndDownload `json:"file"`
}

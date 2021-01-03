package response

import "github.com/lianmi/servers/internal/app/gin-vue-admin/model"

type FilePathResponse struct {
	FilePath string `json:"filePath"`
}

type FileResponse struct {
	File model.ExaFile `json:"file"`
}

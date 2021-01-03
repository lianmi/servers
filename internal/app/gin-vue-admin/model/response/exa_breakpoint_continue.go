package response

import "github.com/lianmi/servers/cmd/gin-vue-admin/model"

type FilePathResponse struct {
	FilePath string `json:"filePath"`
}

type FileResponse struct {
	File model.ExaFile `json:"file"`
}

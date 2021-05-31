/*
这个文件是和前端相关的restful接口-app版本，无须校验
*/
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/common/codes"
	"go.uber.org/zap"
)

//获取服务条款，需要token
func (pc *LianmiApisController) GetFuwutiaokuan(c *gin.Context) {
	code := codes.SUCCESS
	RespData(c, http.StatusOK, code, LMCommon.Fuwutiaokuan) //服务条款
}

//GET 获取app最新版本
func (pc *LianmiApisController) GetAppVersion(c *gin.Context) {
	code := codes.SUCCESS
	oldVersion := c.Query("version")
	os := c.Query("os")

	if oldVersion == "" {
		RespData(c, http.StatusOK, 400, "version is empty")
		return
	}

	if os == "" {
		os = "android"
	}

	type VersionDownloadURL struct {
		Version     string `json:"version"`      // 例如： 8.0.3
		DownloadURL string `json:"download_url"` // 获取app安装包下载链接
	}

	version, err := pc.service.GetAppVersion(oldVersion)
	if err != nil {
		pc.logger.Error("GetAppVersion error", zap.Error(err))
		RespData(c, http.StatusOK, 500, "GetAppVersion error")
		return
	}

	var resp VersionDownloadURL
	if os == "android" {
		resp = VersionDownloadURL{
			Version:     version,
			DownloadURL: LMCommon.ApkDownloadURL,
		}
	} else {
		resp = VersionDownloadURL{
			Version:     version,
			DownloadURL: LMCommon.AppStoreURL,
		}
	}

	RespData(c, http.StatusOK, code, resp)
}

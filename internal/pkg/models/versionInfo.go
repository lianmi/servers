package models

import "github.com/lianmi/servers/internal/pkg/models/global"

//app版本号， X.Y.Z
//如果有大变动，向下不兼容，需要更新X位。
//如果是新增了功能，但是向下兼容，需要更新Y位。
// 如果只是修复bug，需要更新Z位。
type VersionInfo struct {
	Version string `json:"version"` // 例如： 8.0.3
	Details string `json:"details"` // 升级详情，富文本, UI需要用webview显示
}

//app版本号历史表
type AppVersionHistory struct {
	global.LMC_Model

	VersionInfo
}

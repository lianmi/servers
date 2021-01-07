package models

import (
	"github.com/lianmi/servers/internal/pkg/models/global"
)

//定义 Prekey 的数据结构
type Prekey struct {
	global.LMC_Model

	Publickey    string `gorm:"primarykey" form:"publickey" json:"publickey,omitempty"` //公钥
	Type         int    `form:"type" json:"type ,omitempty" `                           //缓存在服务端的(0) - 用于任务类商品加解密 (1)
	Username     string `form:"username" json:"username,omitempty"`                     //用户账号id
	UploadTimeAt int64  `form:"upload_timeat" json:"upload_timeat,omitempty"`           //上传时间
}

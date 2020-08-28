package models

//定义 Prekey 的数据结构
type Prekey struct {
	ID           uint64 `gorm:"primary_key" form:"id" json:"id,omitempty"`    //自动递增id
	Type         int    `form:"type" json:"type ,omitempty" `                 //缓存在服务端的(0) - 用于任务类商品加解密 (1)
	Username     string `form:"username" json:"username,omitempty"`           //用户账号id
	Publickey    string `form:"publickey" json:"publickey,omitempty"`         //公钥
	UploadTimeAt int64  `form:"upload_timeat" json:"upload_timeat,omitempty"` //上传时间
}

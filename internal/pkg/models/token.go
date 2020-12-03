package models

import "time"

//定义token的数据结构
type Token struct {
	Username  string    `gorm:"primarykey"  form:"username" json:"username,omitempty"`
	ExpiredAt time.Time `form:"expired_at" json:"expired_at,omitempty"`     //过期时刻
	Token     string    `gorm:"type:text;not null" json:"token,omitempty" ` // Text文本类型
}

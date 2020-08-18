package models

import "time"

//定义token的数据结构
type Token struct {
	ID        uint64    `gorm:"primary_key" form:"id" json:"id,omitempty"` //自动递增id
	Username  string    `form:"username" json:"username,omitempty"`
	ExpiredAt time.Time `form:"expired_at" json:"expired_at,omitempty"`     //过期时刻
	Token     string    `gorm:"type:text;not null" json:"token,omitempty" ` // Text文本类型
}

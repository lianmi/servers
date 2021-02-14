/*
用户标签表
*/
package models

import (
	"time"

	"github.com/lianmi/servers/internal/pkg/models/global"
	"gorm.io/gorm"
)

//定义标签表的数据结构
type Tag struct {
	global.LMC_Model

	Username string `json:"username" `                          //用户注册号
	TagType  int    `form:"tag_type" json:"tag_type,omitempty"` //标签类型
}

//BeforeCreate CreatedAt赋值
func (t *Tag) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (t *Tag) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

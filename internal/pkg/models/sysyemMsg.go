package models

import (
	"time"

	"github.com/lianmi/servers/internal/pkg/models/global"
	"gorm.io/gorm"
)

//系统公告表

type SystemMsg struct {
	global.LMC_Model

	Level   int    `form:"level" json:"level"`     //公告等级
	Title   string `form:"title" json:"title"`     //标题
	Content string `form:"content" json:"content"` //内容

	Active bool `form:"active" json:"active"` //是否显示

}

//BeforeCreate CreatedAt赋值
func (l *SystemMsg) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (l *SystemMsg) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

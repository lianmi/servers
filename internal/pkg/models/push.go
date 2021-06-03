package models

import (
	"time"

	"github.com/lianmi/servers/internal/pkg/models/global"
	"gorm.io/gorm"
)

//定义用户app的推送设置开关
type PushSetting struct {
	global.LMC_Model
	Username        string `gorm:"username" json:"username" validate:"required"` //用户注册号
	NewRemindSwitch bool   `gorm:"new_remind_switch" json:"new_remind_switch" `  // 新消息提醒
	DetailSwitch    bool   `json:"detail_switch" json:"detail_switch"`           // 通知栏消息详情
	TeamSwitch      bool   `json:"team_switch" json:"team_switch"`               // 群聊消息提醒
	SoundSwitch     bool   `json:"sound_switch" json:"sound_switch"`             // 声音提醒

}

//BeforeCreate CreatedAt赋值
func (l *PushSetting) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (l *PushSetting) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

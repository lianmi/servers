package models

import (
	"github.com/lianmi/servers/internal/pkg/models/global"
	"gorm.io/gorm"
	"time"
)

/*
定义好友表的数据结构
状态:
0-预审核， 1-好友 2-移除好友

只要一加好友，就意味着有两条记录, 删除好友是双向的，也会产生两条记录
*/
type Friend struct {
	global.LMC_Model

	UserID         uint64 `form:"user_id" json:"user_id,omitempty"`                 //用户ID
	FriendUserID   uint64 `form:"friend_user_id" json:"friend_user_id,omitempty"`   //好友的用户ID
	FriendUsername string `form:"friend_username" json:"friend_username,omitempty"` //好友的用户账号
	Alias          string `form:"alias" json:"alias,omitempty"`                     //好友在本地的别名，仅仅自己可见，类似呢称
	Source         string `form:"source" json:"source" binding:"required"`          //好友来源
	Extend         string `form:"extend" json:"extend,omitempty" `                  //扩展字段
	State          int    `form:"state" json:"state,omitempty"`                     //状态， 0-预审核， 1-好友 2-移除好友
	PassAt         int64  `form:"pass_at" json:"pass_at,omitempty"`                 //对方通过好友的时刻
	RejectAt       int64  `form:"reject_at" json:"reject_at,omitempty"`             //对方拒绝加好友的时刻
}

//BeforeCreate CreatedAt赋值
func (d *Friend) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

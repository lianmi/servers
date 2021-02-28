package models

import (
	"time"

	"github.com/lianmi/servers/internal/pkg/models/global"

	"github.com/lianmi/servers/api/proto/team"
	"gorm.io/gorm"
)

/*
CREATE TABLE `tuser` (`teamid` TEXT NOT NULL, `account` TEXT NOT NULL, `avatar` TEXT, `type` INTEGER, `nick` TEXT, `source` TEXT, `mute` INTEGER NOT NULL, `notifytype` INTEGER, `jointime` INTEGER NOT NULL, `ex` TEXT, `createat` INTEGER NOT NULL, `modifyat` INTEGER NOT NULL, PRIMARY KEY(`teamId`, `account`))
*/

//定义群用户的数据结构
type TeamUserInfo struct {
	JoinAt          int64  `form:"join_at" json:"join_at,omitempty"`                     //入群时间，unix时间戳
	TeamID          string `form:"team_id" json:"team_id" binding:"required"`            //群组id， 以team开头
	Teamname        string `form:"team_name" json:"team_name" binding:"required"`        //群组名， 中英文
	Username        string `form:"user_name" json:"user_name,omitempty"`                 //群成员用户账号
	InvitedUsername string `form:"invited_user_name" json:"invited_user_name,omitempty"` //邀请者的用户账号
	Nick            string `form:"nick" json:"nick" binding:"required"`                  //群成员原本的呢称
	AliasName       string `form:"alias_name" json:"alias_name,omitempty"`               //群成员在群里的别名，可以自行修改或管理员修改
	Avatar          string `form:"avatar" json:"avatar,omitempty"`                       //群成员头像
	Label           string `form:"label" json:"label,omitempty" `                        //群成员标签
	Source          string `form:"source" json:"source,omitempty" `                      //群成员来源
	Extend          string `form:"extend" json:"extend,omitempty" `                      //群成员扩展字段
	TeamMemberType  int    `form:"team_member_type" json:"team_member_type,omitempty" `  //群成员类型, 1-待审核的申请加入用户, 2-管理员, 3-普通成员, 4-Owner(群主)
	IsMute          bool   `form:"is_mute" json:"is_mute,omitempty" `                    //是否被禁言
	Mutedays        int    `form:"mute_days" json:"mute_days,omitempty" `                //禁言时长，0表示永久， 以天为单位
	NotifyType      int    `form:"notify_type" json:"notify_type,omitempty" `            //群消息通知类型 1-群全部消息提醒 2-管理员消息提醒 3-联系人提醒 4-所有消息均不提醒
}

type TeamUser struct {
	global.LMC_Model
	TeamUserInfo
}

//BeforeCreate CreatedAt赋值
func (t *TeamUser) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (t *TeamUser) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

func (t *TeamUserInfo) GetType() team.TeamMemberType {
	return team.TeamMemberType(t.TeamMemberType)
}

func (t *TeamUserInfo) GetNotifyType() team.NotifyType {
	return team.NotifyType(t.NotifyType)
}

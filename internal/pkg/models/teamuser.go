package models

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/lianmi/servers/api/proto/team"
)

/*
CREATE TABLE `tuser` (`teamid` TEXT NOT NULL, `account` TEXT NOT NULL, `avatar` TEXT, `type` INTEGER, `nick` TEXT, `source` TEXT, `mute` INTEGER NOT NULL, `notifytype` INTEGER, `jointime` INTEGER NOT NULL, `ex` TEXT, `createat` INTEGER NOT NULL, `modifyat` INTEGER NOT NULL, PRIMARY KEY(`teamId`, `account`))
*/

//定义群用户的数据结构
type TeamUser struct {
	ID             uint64 `form:"id" json:"id,omitempty"`
	JoinAt         int64  `form:"join_at" json:"join_at,omitempty"`                    //入群时间，unix时间戳
	UpdatedAt      int64  `form:"updated_at" json:"updated_at,omitempty"`              //最近更新时间，unix时间戳
	Teamname       string `form:"team_name" json:"team_name" binding:"required"`       //群组id， 以team开头
	Username       string `form:"user_name" json:"user_name,omitempty"`                //群成员用户账号
	Nick           string `form:"nick" json:"nick" binding:"required"`                 //群成员呢称
	Avatar         string `form:"avatar" json:"avatar,omitempty"`                      //群成员头像
	Label          string `form:"label" json:"label,omitempty" `                       //群成员标签
	Source         string `form:"source" json:"source,omitempty" `                     //群成员来源
	Extend         string `form:"extend" json:"extend,omitempty" `                     //群成员扩展字段
	TeamMemberType int    `form:"team_member_type" json:"team_member_type,omitempty" ` //群成员类型
	IsMute         bool   `form:"is_mute" json:"is_mute,omitempty" `                   //是否被禁言
	NotifyType     int    `form:"notify_type" json:"notify_type,omitempty" `           //通知类型
	Province       string `form:"province" json:"province,omitempty" `                 //省份, 如广东省
	City           string `form:"city" json:"city,omitempty" `                         //城市，如广州市
	County         string `form:"county" json:"county,omitempty" `                     //区，如天河区
	Street         string `form:"street" json:"street,omitempty" `                     //街道
	Address        string `form:"address" json:"address,omitempty" `                   //地址
}

//BeforeUpdate UpdatedAt赋值
func (t *TeamUser) BeforeUpdate(scope *gorm.Scope) error {
	scope.SetColumn("UpdatedAt", time.Now().Unix())
	return nil
}

func (t *TeamUser) GetType() team.TeamMemberType {
	return team.TeamMemberType(t.TeamMemberType)
}

func (t *TeamUser) GetNotifyType() team.NotifyType {
	return team.NotifyType(t.NotifyType)
}

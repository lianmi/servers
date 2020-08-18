package models

import (
	"github.com/lianmi/servers/api/proto/team"
)

/*
CREATE TABLE `tuser` (`teamid` TEXT NOT NULL, `account` TEXT NOT NULL, `avatar` TEXT, `type` INTEGER, `nick` TEXT, `source` TEXT, `mute` INTEGER NOT NULL, `notifytype` INTEGER, `jointime` INTEGER NOT NULL, `ex` TEXT, `createat` INTEGER NOT NULL, `modifyat` INTEGER NOT NULL, PRIMARY KEY(`teamId`, `account`))
*/

//定义群用户的数据结构
type TeamUser struct {
	ID                uint64              `form:"id" json:"id,omitempty"`
	UserID            uint64              `form:"user_id" json:"user_id" binding:"required"` //用户id
	Account           string              `form:"account" json:"account,omitempty"`
	Avatar            string              `form:"avatar" json:"avatar,omitempty"`
	Label             string              `form:"label" json:"label,omitempty" `
	Extend            string              `form:"extend" json:"extend,omitempty" `
	MemberShipType    team.TeamMemberType `form:"member_ship_type" json:"member_ship_type,omitempty" ` //群成员类型
	Nick              string              `form:"nick" json:"nick" binding:"required"`
	Source            string              `form:"source" json:"source,omitempty" `
	IsMute            bool                `form:"is_mute" json:"is_mute,omitempty" ` //是否被禁言
	City              string              `form:"city" json:"city,omitempty" `
	County            string              `form:"county" json:"county,omitempty" `
	Street            string              `form:"street" json:"street,omitempty" `
	Address           string              `form:"address" json:"address,omitempty" `
	Branches          string              `form:"branches" json:"branches,omitempty" `                       //网点名称
	LegalPerson       string              `form:"legal_person" json:"legal_person,omitempty" `               //法人姓名
	LegalIdentityCard string              `form:"legal_identity_card" json:"legal_identity_card,omitempty" ` //法人身份证
	CreatedAt         int64              `form:"created_at" json:"created_at,omitempty"`
	UpdatedAt         int64              `form:"updated_at" json:"updated_at,omitempty"`
}

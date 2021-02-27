package models

import (
	"time"

	"github.com/lianmi/servers/internal/pkg/models/global"
	"gorm.io/gorm"

	"github.com/lianmi/servers/api/proto/team"
)

//定义群组的数据结构
type TeamInfo struct {
	TeamID       string `form:"team_id" json:"team_id" `                                           //群id，自动生成，字母(team) + 数字
	Teamname     string `form:"teamname" json:"teamname" `                                         //群名
	Nick         string `json:"nick" validate:"required"`                                          //群呢称，必填
	Icon         string `form:"icon" json:"icon,omitempty"`                                        //群头像url
	Announcement string `form:"announcement" json:"announcement,omitempty" `                       //群公告
	Introductory string `gorm:"type:text;null" form:"introductory" json:"introductory,omitempty" ` // Text文本类型
	Status       int    `form:"status" json:"status"`                                              //状态 Init(1) - 初始状态,未审核 Normal(2) - 正常状态 Blocked(3) - 封禁状态
	Extend       string `form:"extend" json:"extend,omitempty" `                                   //扩展字段
	Owner        string `form:"owner" json:"owner,omitempty" `                                     //群主账号id
	Type         int    `form:"type" json:"type ,omitempty" `                                      //Normal(1) - 普通群 Advanced(2) - vip群
	VerifyType   int    `form:"verify_type" json:"verify_type,omitempty" `                         //1-所有人可加入 2- 需要审核加入 3-仅限邀请加入 4-关注网点后即可入群
	InviteMode   int    `form:"invite_mode" json:"invite_mode,omitempty" `                         //邀请模式 All(1)  Manager(2) Confirm(3)
	MemberLimit  int    `form:"member_limit" json:"member_limit,omitempty" `                       //人数上限
	MemberNum    int    `form:"member_num" json:"member_num,omitempty" `                           //当前成员总数
	MuteType     int    `form:"mute_type" json:"mute_type,omitempty" `                             //禁言类型
	Ex           string `form:"ex" json:"ex,omitempty" `                                           //JSON扩展字段,由业务方解析
	ModifiedBy   string `form:"modified_by" json:"modified_by,omitempty"`                          //最后由哪个操作员修改
	IsMute       bool   `form:"is_mute" json:"is_mute,omitempty"`                                  //群是否被系统禁言
}
type Team struct {
	global.LMC_Model

	TeamInfo
}

//BeforeCreate CreatedAt赋值
func (t *Team) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (t *Team) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

func (t *TeamInfo) GetType() team.TeamMemberType {
	return team.TeamMemberType(t.Type)
}

func (t *TeamInfo) GetVerifyType() team.VerifyType {
	return team.VerifyType(t.VerifyType)
}

func (t *TeamInfo) GetInviteMode() team.InviteMode {
	return team.InviteMode(t.InviteMode)
}

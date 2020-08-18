package models

/*
CREATE TABLE `contact` (`account` TEXT NOT NULL, `alias` TEXT, `source` TEXT, `ex` TEXT, `createat` INTEGER NOT NULL, `modifyat` INTEGER NOT NULL, PRIMARY KEY(`account`))

*/

//定义联系人表的数据结构
type Contact struct {
	ID           uint64 `form:"id" json:"id,omitempty"`
	UserID       uint64 `form:"user_id" json:"user_id" binding:"required"` //用户id
	Account      string `form:"account" json:"account,omitempty"`
	Alias        string `form:"alias" json:"alias,omitempty"`
	Source       string `form:"source" json:"source" binding:"required"` //好友来源
	Extend       string `form:"extend" json:"extend,omitempty" `
	LatestChatAt uint64 `form:"latest_chat_at" json:"latest_chat_at,omitempty""` //最后一次聊天时间(unix)
	CreatedAt    int64 `form:"created_at" json:"created_at,omitempty"`
	UpdatedAt    int64 `form:"updated_at" json:"updated_at,omitempty"`
}

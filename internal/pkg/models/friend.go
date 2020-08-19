package models

/*
定义好友表的数据结构
状态:
0-预审核， 1-好友 2-移除好友 3-禁止加此用户为好友

只要一加好友，就意味着有两条记录
*/
type Friend struct {
	ID           uint64 `gorm:"primary_key" form:"id" json:"id,omitempty"`      //自动递增id
	UserID       uint64 `form:"user_id" json:"user_id,omitempty"`               //用户ID
	FriendUserID uint64 `form:"friend_user_id" json:"friend_user_id,omitempty"` //好友的用户ID
	State        int    `form:"state" json:"state,omitempty"`                   //状态， 0-预审核， 1-好友 2-移除好友 3-禁止加此用户为好友
	CreatedAt    int64  `form:"created_at" json:"created_at,omitempty"`         //创建时刻， 也就是请求加好友的时刻
	PassAt       int64  `form:"pass_at" json:"pass_at,omitempty"`               //对方通过好友的时刻
	RejectAt     int64  `form:"reject_at" json:"reject_at,omitempty"`           //对方拒绝加好友的时刻
	DeleteAt     int64  `form:"delete_at" json:"delete_at,omitempty"`           //删除好友的时刻
}

/*
用户标签表
*/
package models

//定义标签表的数据结构
type Tag struct {
	ID        uint64 `gorm:"primarykey" form:"id" json:"id,omitempty"` //自动递增id
	CreatedAt int64  `form:"created_at" json:"created_at,omitempty"`    //创建时刻
	UpdatedAt int64  `form:"updated_at" json:"updated_at,omitempty"`    //更新时刻
	Username  string `json:"username" `                                 //用户注册号
	TagType   int    `form:"tag_type" json:"tag_type,omitempty"`        //标签类型
}

/*
用户标签表
*/
package models

import (
	"github.com/lianmi/servers/internal/pkg/models/global"
)

//定义标签表的数据结构
type Tag struct {
	global.LMC_Model

	Username string `json:"username" `                          //用户注册号
	TagType  int    `form:"tag_type" json:"tag_type,omitempty"` //标签类型
}

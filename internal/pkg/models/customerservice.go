package models

import (
	"github.com/lianmi/servers/internal/pkg/models/global"
	"gorm.io/gorm"
	"time"
)

/*
客服技术人员表
*/
type CustomerServiceInfo struct {
	global.LMC_Model

	Username string `gorm:"primarykey" json:"username" ` //客服或技术人员的注册账号id

	JobNumber  string `json:"job_number" ` //客服或技术人员的工号
	Type       int    `json:"type" `       //客服或技术人员的类型， 1-客服，2-技术
	Evaluation string `json:"evaluation" ` //职称, 技术工程师，技术员等
	NickName   string `json:"nick_name" `  //呢称
}

//BeforeCreate CreatedAt赋值
func (d *CustomerServiceInfo) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (d *CustomerServiceInfo) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

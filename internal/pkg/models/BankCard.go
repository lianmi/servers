package models

import (
	"time"

	"github.com/lianmi/servers/internal/pkg/models/global"
	"gorm.io/gorm"
)

//定义银行卡的数据结构
type BankCardBase struct {
	Username     string `gorm:"primarykey" json:"username" `                     //用户注册号
	BankTrueName string `form:"bank_true_name" json:"bank_true_name,omitempty" ` //用户实名
	BankCard     string `form:"bank_card" json:"bank_card,omitempty" `           //银行卡账号
	Bank         string `form:"bank" json:"bank,omitempty" `                     //开户银行
}

type BankCard struct {
	global.LMC_Model

	BankCardBase
}

//BeforeCreate CreatedAt赋值
func (user *BankCard) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (user *BankCard) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

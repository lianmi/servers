package models

import (
	"github.com/lianmi/servers/internal/pkg/models/global"
	"gorm.io/gorm"
	"time"
)

/*
会员层级表
一旦用户注册后，就需要新增一条数据
向后的一级One， 向后的二级Two， 向后的三级Three
Three->Two-One->User
*/
type Distribution struct {
	global.LMC_Model

	Username           string `gorm:"primarykey" json:"username"  validate:"required"` //用户注册账号id
	BusinessUsername   string `json:"business_username"  validate:"required"`          //归属的商户注册账号id
	UsernameLevelOne   string `json:"username_level_one" `                             //向后的一级
	UsernameLevelTwo   string `json:"username_level_two" `                             //向后的二级
	UsernameLevelThree string `json:"username_level_three" `                           //向后的三级
}

//BeforeCreate CreatedAt赋值
func (d *Distribution) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (d *Distribution) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

/*
会员付费后佣金分配详情表
一旦归属于某个商户的用户付费，就需要根据BusinessDistribution表的归属商户以及如果向后3级用户账号非空（1，2，3）的话新增佣金数据
Commission:  One:35, Two:10, Three: 5

 alter table `commissions` drop column `year_month`;
*/
type Commission struct {
	global.LMC_Model

	YearMonth        string  `json:"year_month" validate:"required"`        //统计月份
	UsernameLevel    string  `json:"username_level" validate:"required"`    //One Two Three
	BusinessUsername string  `json:"business_username" validate:"required"` //归属的商户注册账号id
	Amount           float64 `json:"amount" validate:"required"`            //会员费用金额，单位是人民币
	OrderID          string  `json:"order_id" validate:"required"`          //订单ID
	Commission       float64 `json:"commission" validate:"required"`        //本次佣金提成金额，人民币
}

//BeforeCreate CreatedAt赋值
func (c *Commission) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (c *Commission) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

/*
用户的佣金月统计,  每个用户每月生成一条记录
利用复合主键（联合主键） username 及 year_month 控制Save方法, 这条数据如果在数据库中存在，就做更新操作；如果不存在就做插入操作。
*/
type CommissionStatistics struct {
	global.LMC_Model

	Username        string  `gorm:"primarykey" json:"username" validate:"required"`   //用户户注册账号id
	YearMonth       string  `gorm:"primarykey" json:"year_month" validate:"required"` //统计月份
	TotalCommission float64 `json:"total_commission" validate:"required"`             //本月佣金合计
	IsRebate        bool    `json:"is_rebate" `                                       //是否支付了佣金
	RebateTime      int64   `form:"rebate_time" json:"rebate_time,omitempty"`         //平台操作返佣时间
}

//BeforeCreate CreatedAt赋值
func (n *CommissionStatistics) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (n *CommissionStatistics) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

/*
商户所属会员详情表
一旦归属于某个商户的用户被推荐注册，就增加一条记录
*/
type BusinessUnderling struct {
	global.LMC_Model

	MembershipUsername string `json:"membership_username" validate:"required"` //会员账户
	BusinessUsername   string `json:"business_username" validate:"required"`   //归属的商户注册账号id
}

//BeforeCreate CreatedAt赋值
func (bc *BusinessUnderling) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (bc *BusinessUnderling) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

/*
商户下属的月会员统计表
每月一条记录
(统计BusinessUnderling表的某个商户的每月会员的数量)
*/
type BusinessUserStatistics struct {
	global.LMC_Model

	BusinessUsername string `json:"business_username" validate:"required"` //商户注册账号id
	YearMonth        string `json:"year_month" validate:"required"`        //统计月份
	UnderlingTotal   int64  `json:"underling_total" validate:"required"`   //下属用户总数
}

//BeforeCreate CreatedAt赋值
func (b *BusinessUserStatistics) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (b *BusinessUserStatistics) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

/*
佣金提现申请表 CommissionWithdraw
佣金由经过申请，由平台转账到用户或商户钱包
*/
type CommissionWithdraw struct {
	global.LMC_Model

	Username           string  `json:"username" validate:"required"`                  //用户或商户注册账号id
	UserType           int     `form:"user_type" json:"user_type" binding:"required"` //用户类型 1-普通，2-商户
	YearMonth          string  `json:"year_month" validate:"required"`                //统计月份
	WithdrawCommission float64 `json:"withdraw_commission,omitempty"`                 //佣金提现金额
	IsConfirm          bool    `json:"is_confirm,omitempty" `                         //审核是否通过
	OpUsername         string  `json:"op_username,omitempty"`                         //操作员账号，谁审核
}

//BeforeCreate CreatedAt赋值
func (bw *CommissionWithdraw) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (bw *CommissionWithdraw) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

/*
会员层级表
一旦用户注册后，就需要新增一条数据
向后的一级One， 向后的二级Two， 向后的三级Three
Three->Two-One->User
*/
type Distribution struct {
	ID                 uint64 `gorm:"primary_key" form:"id" json:"id,omitempty"` //自动递增id
	CreatedAt          int64  `form:"created_at" json:"created_at,omitempty"`    //创建时刻,毫秒
	UpdatedAt          int64  `form:"updated_at" json:"updated_at,omitempty"`    //修改时间
	Username           string `json:"username"  validate:"required"`             //用户注册账号id
	BusinessUsername   string `json:"business_username"  validate:"required"`    //归属的商户注册账号id
	UsernameLevelOne   string `json:"username_level_one" `                       //向后的一级
	UsernameLevelTwo   string `json:"username_level_two" `                       //向后的二级
	UsernameLevelThree string `json:"username_level_three" `                     //向后的三级
}

//BeforeCreate CreatedAt赋值
func (d *Distribution) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (d *Distribution) BeforeUpdate(scope *gorm.Scope) error {
	scope.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

/*
会员付费后佣金分配详情表
一旦归属于某个商户的用户付费，就需要根据BusinessDistribution表的归属商户以及如果向后3级用户账号非空（1，2，3）的话新增佣金数据
UsCommissioner:  One:35, Two:10, Three: 5
*/
type Commission struct {
	ID               uint64  `gorm:"primary_key" form:"id" json:"id,omitempty"`                   //自动递增id
	CreatedAt        int64   `form:"created_at" json:"created_at,omitempty"`                      //创建时刻,毫秒
	UpdatedAt        int64   `form:"updated_at" json:"updated_at,omitempty"`                      //修改时间
	UsernameLevel    string  `json:"username_level" validate:"required"`                          //One Two Three
	BusinessUsername string  `json:"business_username" validate:"required"`                       //归属的商户注册账号id
	Amount           float64 `json:"amount" validate:"required"`                                  //会员费用金额，单位是人民币
	OrderID          string  `json:"order_id" validate:"required"`                                //订单ID
	BlockNumber      uint64  `json:"block_number" validate:"required"`                            //交易成功打包的区块高度
	TxHash           string  `json:"tx_hash" validate:"required"`                                 //交易成功打包的区块哈希
	Content          string  `gorm:"type:longtext;null" form:"content" json:"content,omitempty" ` //附言 Text文本类型
	Commission       float64 `json:"commission" validate:"required"`                              //本次佣金提成5 10 35
}

//BeforeCreate CreatedAt赋值
func (c *Commission) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (c *Commission) BeforeUpdate(scope *gorm.Scope) error {
	scope.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

/*
普通用户的佣金月统计, 每日凌晨4点由定时任务生成(一旦返佣或IsClose=true不再统计), 用户佣金由经过用户申请，由平台转账到用户钱包
当redis里的时间戳更改才统计，这样就节省资源
*/
type NormalUserCommissionStatistics struct {
	ID              uint64  `gorm:"primary_key" form:"id" json:"id,omitempty"` //自动递增id
	CreatedAt       int64   `form:"created_at" json:"created_at,omitempty"`    //创建时刻,毫秒
	UpdatedAt       int64   `form:"updated_at" json:"updated_at,omitempty"`    //修改时间
	Username        string  `json:"username" validate:"required"`              //用户户注册账号id
	YearMonth       string  `json:"year_month" validate:"required"`            //统计月份
	TotalCommission float64 `json:"total_commission" validate:"required"`      //本月佣金合计
	IsClose         bool    `json:"is_close" `                                 //是否关闭自动统计，true-已关闭， 不会自动统计
	IsRebate        bool    `json:"is_rebate" `                                //是否支付了佣金
}

//BeforeCreate CreatedAt赋值
func (n *NormalUserCommissionStatistics) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (n *NormalUserCommissionStatistics) BeforeUpdate(scope *gorm.Scope) error {
	scope.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

/*
商户所属会员详情表
一旦归属于某个商户的用户付费，就增加一条记录
*/
type BusinessCommission struct {
	ID                 uint64  `gorm:"primary_key" form:"id" json:"id,omitempty"`                   //自动递增id
	CreatedAt          int64   `form:"created_at" json:"created_at,omitempty"`                      //创建时刻,毫秒
	UpdatedAt          int64   `form:"updated_at" json:"updated_at,omitempty"`                      //修改时间
	MembershipUsername string  `json:"membership_username" validate:"required"`                     //缴费的会员账户
	BusinessUsername   string  `json:"business_username" validate:"required"`                       //归属的商户注册账号id
	Amount             float64 `json:"amount" validate:"required"`                                  //会员费用金额，单位是人民币
	OrderID            string  `json:"order_id" validate:"required"`                                //订单ID
	BlockNumber        uint64  `json:"block_number" validate:"required"`                            //交易成功打包的区块高度
	TxHash             string  `json:"tx_hash" validate:"required"`                                 //交易成功打包的区块哈希
	Content            string  `gorm:"type:longtext;null" form:"content" json:"content,omitempty" ` //附言 Text文本类型
	Commission         float64 `json:"commission" validate:"required"`                              //本次佣金提成, 默认是 11元
}

//BeforeCreate CreatedAt赋值
func (bc *BusinessCommission) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (bc *BusinessCommission) BeforeUpdate(scope *gorm.Scope) error {
	scope.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

/*
商户的所属月统计(统计BusinessCommission表), 一旦返佣后就不统计, 商户佣金由经过商户申请，由平台转账到商户钱包
当redis里的时间戳更改才统计，这样就节省资源
*/
type BusinessUserCommissionStatistics struct {
	ID               uint64  `gorm:"primary_key" form:"id" json:"id,omitempty"` //自动递增id
	CreatedAt        int64   `form:"created_at" json:"created_at,omitempty"`    //创建时刻,毫秒
	UpdatedAt        int64   `form:"updated_at" json:"updated_at,omitempty"`    //修改时间
	BusinessUsername string  `json:"business_username" validate:"required"`     //商户注册账号id
	YearMonth        string  `json:"year_month" validate:"required"`            //统计月份
	Total            int64   `json:"total" validate:"required"`                 //下属用户总数
	TotalCommission  float64 `json:"total_commission" validate:"required"`      //本月佣金合计
	IsRebate         bool    `json:"is_rebate" `                                //是否支付了佣金
}

//BeforeCreate CreatedAt赋值
func (b *BusinessUserCommissionStatistics) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (b *BusinessUserCommissionStatistics) BeforeUpdate(scope *gorm.Scope) error {
	scope.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

package models

import (
	"github.com/lianmi/servers/internal/pkg/models/global"
	"gorm.io/gorm"
	"time"
)

/*
VIP会员价格表及优惠设定，计算分佣时需要考虑是否有折扣
*/
type VipPrice struct {
	global.LMC_Model

	PayType          int     `json:"pay_type" form:"payType" gorm:"column:pay_type;comment:VIP类型，1-包年，2-包季， 3-包月"`
	Price            float32 `json:"price" form:"price" gorm:"column:price;comment:价格, 单位: 元"`
	BusinessUsername string  `json:"business_username" form:"business_username" gorm:"column:business_username;comment:购买会员的商户账号，一般都是预先注册，然后再写死;type:varchar(191);size:191;"`
	ProductID        string  `json:"product_id" form:"product_id" gorm:"column:product_id;comment:商品ID;type:varchar(191);size:191;"`
	Title            string  `json:"title" form:"title" gorm:"column:title;comment:价格标题说明;type:varchar(191);size:191;"`
	Days             int     `json:"days" form:"days" gorm:"column:days;comment:开通时长 本记录对应的天数，例如包年增加365天，包季是90天，包月是30天"`
	IsActive         bool    `json:"is_active" form:"is_active" gorm:"column:is_active;comment:此价格是否激活，true的状态才可用"`
}

func (VipPrice) TableName() string {
	return "vip_prices"
}

//BeforeCreate CreatedAt赋值
func (user *VipPrice) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (user *VipPrice) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

/*
系统电子优惠券表 - 第一版只对VIP 7天体验卡 或 抵消手续费有效
VIP 7天体验卡 - 一个用户只能体验一次，体验过之后就算领了卡也无法冲进去
用于抵消手续费的数量，不得用于提现,不存放区块链里，只存放在redis里, 每次只能领取一张
*/
type ECoupon struct {
	global.LMC_Model

	Title     string  `form:"title" json:"title,omitempty" `           //标题说明
	ScopeType int     `form:"scope_type" json:"scope_type,omitempty" ` //作用范围类型,0-默认，不作用于任何收费， 1-会员购买及服务费，2-VIP天数，用于几天VIP的体验卡
	Amount    float64 `form:"amount" json:"amount,omitempty"`          //用于扣除手续费的人民币数量，不得用于提现,不存放区块链里，只存放在redis里
	Days      int     `form:"days" json:"days,omitempty"`              //体验时长天数, 最多不得多于7天
	IsUsed    bool    `form:"is_used" json:"is_used"`                  //是否已经使用,一旦激活就不能再次使用
	Expire    int64   `form:"expire" json:"expire,omitempty"`          //过期时间，0-无限
}

func (ECoupon) TableName() string {
	return "e_coupons"
}

//BeforeCreate CreatedAt赋值
func (user *ECoupon) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (user *ECoupon) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

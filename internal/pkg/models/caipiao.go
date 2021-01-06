package models

import (
	"github.com/lianmi/servers/internal/pkg/models/global"
	"gorm.io/gorm"
	"time"
)

//各个彩种的销售开始及最后截止时间
type LotterySaleTime struct {
	global.LMC_Model

	// LotteryType   int    `form:"lottery_type" json:"lottery_type"`       //彩票种类
	// LotteryName   string `form:"lottery_name" json:"lottery_name"`       //彩票名称
	// SaleStartWeek string `form:"sale_start_week" json:"sale_start_week"` //销售开始 是 星期几，用逗号隔开, 如: 2, 4, 7
	// SaleStartTime string `form:"sale_start_time" json:"sale_start_time"` //销售开始时间, 如:  21:00
	// SaleEndWeek   string `form:"sale_end_week" json:"sale_end_week"`     //销售结束 是 星期几，用逗号隔开, 如: 2, 4, 7
	// SaleEndTime   string `form:"sale_end_time" json:"sale_end_time"`     //销售结束时间,  如:  20:00
	// Holidays      string `form:"holidays" json:"holidays,omitempty"`     //节假日时间，用逗号隔开， 01-01,10-01,etc...
	// IsActive      bool   `form:"is_active" json:"is_active"`             //是否激活, true-激活 ， false  - 不激活
	LotteryType   int    `json:"lotteryType" form:"lotteryType" gorm:"column:lottery_type;comment:;type:bigint;size:19;"`
	LotteryName   string `json:"lotteryName" form:"lotteryName" gorm:"column:lottery_name;comment:;type:varchar(191);size:191;"`
	SaleStartWeek string `json:"saleStartWeek" form:"saleStartWeek" gorm:"column:sale_start_week;comment:星期几开始， 以半角逗号隔开;type:varchar(191);size:191;"`
	SaleStartTime string `json:"saleStartTime" form:"saleStartTime" gorm:"column:sale_start_time;comment:开售时分秒, 22:0:00;type:varchar(191);size:191;"`
	SaleEndWeek   string `json:"saleEndWeek" form:"saleEndWeek" gorm:"column:sale_end_week;comment:停售星期几，以半角逗号隔开 ;type:varchar(191);size:191;"`
	SaleEndTime   string `json:"saleEndTime" form:"saleEndTime" gorm:"column:sale_end_time;comment:停售时分秒, 22:0:00;type:varchar(191);size:191;"`
	Holidays      string `json:"holidays" form:"holidays" gorm:"column:holidays;comment:;type:varchar(191);size:191;"`
	IsActive      *bool  `json:"isActive" form:"isActive" gorm:"column:is_active;comment:;type:tinyint;"`
}

func (LotterySaleTime) TableName() string {
	return "lottery_sale_times"
}

//BeforeCreate CreatedAt赋值
func (d *LotterySaleTime) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate ModifyAt赋值
func (d *LotterySaleTime) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("ModifyAt", time.Now().UnixNano()/1e6)
	return nil
}

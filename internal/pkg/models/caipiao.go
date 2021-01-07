package models

import (
	"github.com/lianmi/servers/internal/pkg/models/global"
	"gorm.io/gorm"
	"time"
)

//各个彩种的销售开始及最后截止时间
type LotterySaleTime struct {
	global.LMC_Model

	LotteryType   int    `json:"lotteryType" form:"lotteryType" gorm:"column:lottery_type;comment:彩票种类;type:bigint;size:19;"`                   //彩票种类
	LotteryName   string `json:"lotteryName" form:"lotteryName" gorm:"column:lottery_name;comment:彩票名称;type:varchar(191);size:191;"`            //彩票名称
	SaleEndWeek   string `json:"saleEndWeek" form:"saleEndWeek" gorm:"column:sale_end_week;comment:停售星期几，以半角逗号隔开 ;type:varchar(191);size:191;"` //销售结束 是 星期几，用逗号隔开, 如: 2, 4, 7
	SaleEndHour   int    `json:"saleEndHour" form:"saleEndHour" gorm:"column:sale_end_hour;comment:停售hour时;type:bigint;size:19;"`               //销售结束hour,  如:  20
	SaleEndMinute int    `json:"saleEndMinute" form:"saleEndMinute" gorm:"column:sale_end_minute;comment:停售minute分;type:bigint;size:19;"`       //销售结束minute,  如:  0s
	Holidays      string `json:"holidays" form:"holidays" gorm:"column:holidays;comment:节假日时间;type:varchar(191);size:191;"`                     //节假日时间，用逗号隔开， 01-01,10-01,etc...
	IsActive      bool   `json:"isActive" form:"isActive" gorm:"column:is_active;comment:是否激活;type:tinyint;"`                                   //是否激活, true-激活 ， false  - 不激活
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

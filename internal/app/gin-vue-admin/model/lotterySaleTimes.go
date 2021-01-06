// 自动生成模板LotterySaleTimes
package model

import (
	// "github.com/lianmi/servers/internal/app/gin-vue-admin/global"
)
/*
// 如果含有time.Time 请自行import time包
type LotterySaleTimes struct {
      global.GVA_MODEL
      LotteryType  int `json:"lotteryType" form:"lotteryType" gorm:"column:lottery_type;comment:;type:bigint;size:19;"`
      LotteryName  string `json:"lotteryName" form:"lotteryName" gorm:"column:lottery_name;comment:;type:varchar(191);size:191;"`
      SaleStartWeek  string `json:"saleStartWeek" form:"saleStartWeek" gorm:"column:sale_start_week;comment:星期几开始， 以半角逗号隔开;type:varchar(191);size:191;"`
      SaleStartTime  string `json:"saleStartTime" form:"saleStartTime" gorm:"column:sale_start_time;comment:开售时分秒, 22:0:00;type:varchar(191);size:191;"`
      SaleEndWeek  string `json:"saleEndWeek" form:"saleEndWeek" gorm:"column:sale_end_week;comment:停售星期几，以半角逗号隔开 ;type:varchar(191);size:191;"`
      SaleEndTime  string `json:"saleEndTime" form:"saleEndTime" gorm:"column:sale_end_time;comment:停售时分秒, 22:0:00;type:varchar(191);size:191;"`
      Holidays  string `json:"holidays" form:"holidays" gorm:"column:holidays;comment:;type:varchar(191);size:191;"`
      IsActive  *bool `json:"isActive" form:"isActive" gorm:"column:is_active;comment:;type:tinyint;"`
}


func (LotterySaleTimes) TableName() string {
  return "lottery_sale_times"
}

*/

// 如果使用工作流功能 需要打开下方注释 并到initialize的workflow中进行注册 且必须指定TableName
// type LotterySaleTimesWorkflow struct {
// 	// 工作流操作结构体
// 	WorkflowBase      `json:"wf"`
// 	LotterySaleTimes   `json:"business"`
// }

// func (LotterySaleTimes) TableName() string {
// 	return "lottery_sale_times"
// }

// 工作流注册代码

// initWorkflowModel内部注册
// model.WorkflowBusinessStruct["lotterySaleTimes"] = func() model.GVA_Workflow {
//   return new(model.LotterySaleTimesWorkflow)
// }

// initWorkflowTable内部注册
// model.WorkflowBusinessTable["lotterySaleTimes"] = func() interface{} {
// 	return new(models.LotterySaleTime)
// }

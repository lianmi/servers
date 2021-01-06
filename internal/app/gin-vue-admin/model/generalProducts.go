// 自动生成模板GeneralProduct
package model

import (
// "github.com/lianmi/servers/internal/app/gin-vue-admin/global"
)

/*
// 如果含有time.Time 请自行import time包
type GeneralProduct struct {
	global.GVA_MODEL
	AllowCancel *bool  `json:"allowCancel" form:"allowCancel" gorm:"column:allow_cancel;comment:;type:tinyint;"`
	DescPic1    string `json:"descPic1" form:"descPic1" gorm:"column:desc_pic1;comment:;type:varchar(191);size:191;"`
	DescPic2    string `json:"descPic2" form:"descPic2" gorm:"column:desc_pic2;comment:;type:varchar(191);size:191;"`
	DescPic3    string `json:"descPic3" form:"descPic3" gorm:"column:desc_pic3;comment:;type:varchar(191);size:191;"`
	DescPic4    string `json:"descPic4" form:"descPic4" gorm:"column:desc_pic4;comment:;type:varchar(191);size:191;"`
	DescPic5    string `json:"descPic5" form:"descPic5" gorm:"column:desc_pic5;comment:;type:varchar(191);size:191;"`
	DescPic6    string `json:"descPic6" form:"descPic6" gorm:"column:desc_pic6;comment:;type:varchar(191);size:191;"`
	// ModifyAt  int `json:"modifyAt" form:"modifyAt" gorm:"column:modify_at;comment:;type:bigint;size:19;"`
	ProductDesc      string `json:"productDesc" form:"productDesc" gorm:"column:product_desc;comment:;type:varchar(191);size:191;"`
	ProductId        string `json:"productId" form:"productId" gorm:"column:product_id;comment:;type:varchar(191);size:191;"`
	ProductName      string `json:"productName" form:"productName" gorm:"column:product_name;comment:;type:varchar(191);size:191;"`
	ProductPic1Large string `json:"productPic1Large" form:"productPic1Large" gorm:"column:product_pic1_large;comment:;type:varchar(191);size:191;"`
	ProductPic2Large string `json:"productPic2Large" form:"productPic2Large" gorm:"column:product_pic2_large;comment:;type:varchar(191);size:191;"`
	ProductPic3Large string `json:"productPic3Large" form:"productPic3Large" gorm:"column:product_pic3_large;comment:;type:varchar(191);size:191;"`
	ProductType      int    `json:"productType" form:"productType" gorm:"column:product_type;comment:"`
	ShortVideo       string `json:"shortVideo" form:"shortVideo" gorm:"column:short_video;comment:;type:varchar(191);size:191;"`
}

func (GeneralProduct) TableName() string {
	return "general_products"
}
*/

// 如果使用工作流功能 需要打开下方注释 并到initialize的workflow中进行注册 且必须指定TableName
// type GeneralProductWorkflow struct {
// 	// 工作流操作结构体
// 	WorkflowBase      `json:"wf"`
// 	GeneralProduct   `json:"business"`
// }

// func (GeneralProduct) TableName() string {
// 	return "general_products"
// }

// 工作流注册代码

// initWorkflowModel内部注册
// model.WorkflowBusinessStruct["generalProducts"] = func() model.GVA_Workflow {
//   return new(model.GeneralProductWorkflow)
// }

// initWorkflowTable内部注册
// model.WorkflowBusinessTable["generalProducts"] = func() interface{} {
// 	return new(model.GeneralProduct)
// }

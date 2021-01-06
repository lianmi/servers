package models

import (
	"time"

	"github.com/lianmi/servers/internal/pkg/models/global"
	"gorm.io/gorm"
)

/*
服务端的通用商品表
*/
type GeneralProduct struct {
	global.LMC_Model

	// ProductID        string `gorm:"primarykey" form:"product_id" json:"product_id,omitempty"` //商品ID
	// ProductName      string `form:"product_name" json:"product_name,omitempty"`               //商品名称
	// ProductType      int    `form:"product_type" json:"product_type,omitempty"`               //商品种类枚举
	// ProductDesc      string `form:"product_desc" json:"product_desc,omitempty"`               //商品详细介绍
	// ProductPic1Large string `form:"product_pic1_large" json:"product_pic1_large,omitempty"`   //商品图片1-大图
	// ProductPic2Large string `form:"product_pic2_large" json:"product_pic2_large,omitempty"`   //商品图片2-大图
	// ProductPic3Large string `form:"product_pic3_large" json:"product_pic3_large,omitempty"`   //商品图片3-大图
	// ShortVideo string `form:"short_video" json:"short_video,omitempty"` //商品短视频
	// AllowCancel bool `form:"allow_cancel" json:"allow_cancel" binding:"required"` //是否允许撤单， 默认是可以，彩票类的不可以

	// DescPic1 string `form:"desc_pic1" json:"desc_pic1,omitempty"`  //商品介绍pic1 -图片1
	// DescPic2 string `form:"desc_pic2" json:"desc_pic2,omitempty"`  //商品介绍pic2 -图片2
	// DescPic3 string `form:"desc_pic3" json:"desc_pic3,omitempty"`  //商品介绍pic3 -图片3
	// DescPic4 string `form:"desc_pic4" json:"desc_pic4,omitempty"`  //商品介绍pic4 -图片4
	// DescPic5 string `form:"desc_pic5" json:"desc_pic5,omitempty"`  //商品介绍pic5 -图片5
	// DescPic6 string `form:"desc_pic6" json:"desc_pic16,omitempty"` //商品介绍pic6 -图片6

	ProductId        string `json:"productId" form:"productId" gorm:"column:product_id;comment:;type:varchar(191);size:191;"`
	ProductName      string `json:"productName" form:"productName" gorm:"column:product_name;comment:;type:varchar(191);size:191;"`
	ProductType      int    `json:"productType" form:"productType" gorm:"column:product_type;comment:"`
	ProductDesc      string `json:"productDesc" form:"productDesc" gorm:"column:product_desc;comment:;type:varchar(191);size:191;"`
	ProductPic1Large string `json:"productPic1Large" form:"productPic1Large" gorm:"column:product_pic1_large;comment:;type:varchar(191);size:191;"`
	ProductPic2Large string `json:"productPic2Large" form:"productPic2Large" gorm:"column:product_pic2_large;comment:;type:varchar(191);size:191;"`
	ProductPic3Large string `json:"productPic3Large" form:"productPic3Large" gorm:"column:product_pic3_large;comment:;type:varchar(191);size:191;"`
	ShortVideo       string `json:"shortVideo" form:"shortVideo" gorm:"column:short_video;comment:;type:varchar(191);size:191;"`
	AllowCancel      *bool  `json:"allowCancel" form:"allowCancel" gorm:"column:allow_cancel;comment:;type:tinyint;"`

	DescPic1 string `json:"descPic1" form:"descPic1" gorm:"column:desc_pic1;comment:;type:varchar(191);size:191;"`
	DescPic2 string `json:"descPic2" form:"descPic2" gorm:"column:desc_pic2;comment:;type:varchar(191);size:191;"`
	DescPic3 string `json:"descPic3" form:"descPic3" gorm:"column:desc_pic3;comment:;type:varchar(191);size:191;"`
	DescPic4 string `json:"descPic4" form:"descPic4" gorm:"column:desc_pic4;comment:;type:varchar(191);size:191;"`
	DescPic5 string `json:"descPic5" form:"descPic5" gorm:"column:desc_pic5;comment:;type:varchar(191);size:191;"`
	DescPic6 string `json:"descPic6" form:"descPic6" gorm:"column:desc_pic6;comment:;type:varchar(191);size:191;"`
}

//BeforeCreate CreatedAt赋值
func (d *GeneralProduct) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate ModifyAt赋值
func (d *GeneralProduct) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("ModifyAt", time.Now().UnixNano()/1e6)
	return nil
}

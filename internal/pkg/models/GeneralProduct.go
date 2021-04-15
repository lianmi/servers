package models

import (
	"time"

	"gorm.io/gorm"
)

/*
服务端的通用商品表
*/
type GeneralProductInfo struct {
	ProductId        string `json:"productId" form:"productId" gorm:"column:product_id;comment:商品UUID;type:varchar(191);size:191;"`
	ProductName      string `json:"productName" form:"productName" gorm:"column:product_name;comment:商品名称;type:varchar(191);size:191;"`
	ProductType      int    `json:"productType" form:"productType" gorm:"column:product_type;comment:商品种类枚举"`
	ProductDesc      string `json:"productDesc" form:"productDesc" gorm:"column:product_desc;comment:商品详细介绍;type:varchar(191);size:191;"`
	ProductPic1Large string `json:"productPic1Large" form:"productPic1Large" gorm:"column:product_pic1_large;comment:商品图片1-大图;type:varchar(191);size:191;"`
	ProductPic2Large string `json:"productPic2Large" form:"productPic2Large" gorm:"column:product_pic2_large;comment:商品图片2-大图;type:varchar(191);size:191;"`
	ProductPic3Large string `json:"productPic3Large" form:"productPic3Large" gorm:"column:product_pic3_large;comment:商品图片3-大图;type:varchar(191);size:191;"`
	ShortVideo       string `json:"shortVideo" form:"shortVideo" gorm:"column:short_video;comment:商品短视频;type:varchar(191);size:191;"`
	AllowCancel      *bool  `json:"allowCancel" form:"allowCancel" gorm:"column:allow_cancel;comment:是否允许撤单， 默认是可以，彩票类的不可以;type:tinyint;"`

	DescPic1 string `json:"descPic1" form:"descPic1" gorm:"column:desc_pic1;comment:商品介绍pic1 -图片1;type:varchar(191);size:191;"`
	DescPic2 string `json:"descPic2" form:"descPic2" gorm:"column:desc_pic2;comment:商品介绍pic2 -图片2;type:varchar(191);size:191;"`
	DescPic3 string `json:"descPic3" form:"descPic3" gorm:"column:desc_pic3;comment:商品介绍pic3 -图片3;type:varchar(191);size:191;"`
	DescPic4 string `json:"descPic4" form:"descPic4" gorm:"column:desc_pic4;comment:商品介绍pic4 -图片4;type:varchar(191);size:191;"`
	DescPic5 string `json:"descPic5" form:"descPic5" gorm:"column:desc_pic5;comment:商品介绍pic5 -图片5;type:varchar(191);size:191;"`
	DescPic6 string `json:"descPic6" form:"descPic6" gorm:"column:desc_pic6;comment:商品介绍pic6 -图片6;type:varchar(191);size:191;"`
}

type GeneralProduct struct {
	//global.LMC_Model
	ID        uint           `gorm:"primarykey"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	GeneralProductInfo
	UpdatedAtInt int64 `json:"updated_at" gorm:"-"`
}

//BeforeCreate CreatedAt赋值
func (d *GeneralProduct) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (d *GeneralProduct) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

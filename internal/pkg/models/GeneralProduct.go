package models

import (
	"time"

	"gorm.io/gorm"
)

/*
服务端的通用商品表
*/
type GeneralProduct struct {
	ProductID        string `gorm:"primarykey" form:"product_id" json:"product_id,omitempty"` //商品ID
	ProductName      string `form:"product_name" json:"product_name,omitempty"`               //商品名称
	ProductType      int    `form:"product_type" json:"product_type,omitempty"`               //商品种类枚举
	ProductDesc      string `form:"product_desc" json:"product_desc,omitempty"`               //商品详细介绍
	ProductPic1Large string `form:"product_pic1_large" json:"product_pic1_large,omitempty"`   //商品图片1-大图
	ProductPic2Large string `form:"product_pic2_large" json:"product_pic2_large,omitempty"`   //商品图片2-大图
	ProductPic3Large string `form:"product_pic3_large" json:"product_pic3_large,omitempty"`   //商品图片3-大图
	// Thumbnail        string `form:"thumbnail" json:"thumbnail,omitempty"`                   //商品短视频缩略图
	ShortVideo  string `form:"short_video" json:"short_video,omitempty"`            //商品短视频
	CreatedAt   int64  `form:"created_at" json:"created_at,omitempty"`              //创建时刻， 也就是上架时刻
	ModifyAt    int64  `form:"modify_at" json:"modify_at,omitempty"`                //最后修改时间
	AllowCancel bool   `form:"allow_cancel" json:"allow_cancel" binding:"required"` //是否允许撤单， 默认是可以，彩票类的不可以

	DescPic1 string `form:"desc_pic1" json:"desc_pic1,omitempty"`  //商品介绍pic1 -图片1
	DescPic2 string `form:"desc_pic2" json:"desc_pic2,omitempty"`  //商品介绍pic2 -图片2
	DescPic3 string `form:"desc_pic3" json:"desc_pic3,omitempty"`  //商品介绍pic3 -图片3
	DescPic4 string `form:"desc_pic4" json:"desc_pic4,omitempty"`  //商品介绍pic4 -图片4
	DescPic5 string `form:"desc_pic5" json:"desc_pic5,omitempty"`  //商品介绍pic5 -图片5
	DescPic6 string `form:"desc_pic6" json:"desc_pic16,omitempty"` //商品介绍pic6 -图片6

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

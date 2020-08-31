package models

/*
服务端的商品表
缓存商户的上架商品
*/
type Product struct {
	ID                uint64  `gorm:"primary_key" form:"id" json:"id,omitempty"`              //自动递增id
	Username          string  `form:"username" json:"username,omitempty"`                     //商户用户账号id
	ProductID         string  `form:"product_id" json:"product_id,omitempty"`                 //商品ID
	ProductName       string  `form:"product_name" json:"product_name,omitempty"`             //商品名称
	CategoryName      string  `form:"category_name" json:"category_name,omitempty"`           //商品分类名称
	ProductDesc       string  `form:"product_desc" json:"product_desc,omitempty"`             //商品详细介绍
	ProductPic1       string  `form:"product_pic1" json:"product_pic1,omitempty"`             //商品图片1
	ProductPic2       string  `form:"product_pic2" json:"product_pic2,omitempty"`             //商品图片2
	ProductPic3       string  `form:"product_pic3" json:"product_pic3,omitempty"`             //商品图片3
	ProductPic4       string  `form:"product_pic4" json:"product_pic4,omitempty"`             //商品图片4
	ProductPic5       string  `form:"product_pic5" json:"product_pic5,omitempty"`             //商品图片5
	ShortVideo1       string  `form:"short_video1" json:"short_video1,omitempty"`             //商品短视频1
	ShortVideo2       string  `form:"short_video2" json:"short_video2,omitempty"`             //商品短视频2
	ShortVideo3       string  `form:"short_video3" json:"short_video3,omitempty"`             //商品短视频3
	Price             float32 `form:"price" json:"price,omitempty"`                           //价格
	LeftCount         uint64  `form:"left_count" json:"left_count,omitempty"`                 //库存数量
	Discount          float32 `form:"discount" json:"discount,omitempty"`                     //折扣
	DiscountDesc      string  `form:"discount_desc" json:"discount_desc,omitempty"`           //折扣说明
	DiscountStartTime int64   `form:"discount_starttime" json:"discount_starttime,omitempty"` //折扣开始时间
	DiscountEndTime   int64   `form:"discount_endtime" json:"discount_endtime,omitempty"`     //折扣结束时间
	CreateAt          int64   `form:"create_at" json:"create_at,omitempty"`                 //创建时刻， 也就是上架时刻
	ModifyAt          int64   `form:"modify_at" json:"modify_at,omitempty"`                   //最后修改时间
}

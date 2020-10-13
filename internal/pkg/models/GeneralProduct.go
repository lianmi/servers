package models

/*
服务端的通用商品表
*/
type GeneralProduct struct {
	ID                uint64 `gorm:"primary_key" form:"id" json:"id,omitempty"`                //自动递增id
	ProductID         string `form:"product_id" json:"product_id,omitempty"`                   //商品ID
	ProductName       string `form:"product_name" json:"product_name,omitempty"`               //商品名称
	CategoryName      string `form:"category_name" json:"category_name,omitempty"`             //商品分类名称
	ProductDesc       string `form:"product_desc" json:"product_desc,omitempty"`               //商品详细介绍
	ProductPic1Small  string `form:"product_pic1_small" json:"product_pic1_small,omitempty"`   //商品图片1-小图
	ProductPic1Middle string `form:"product_pic1_middle" json:"product_pic1_middle,omitempty"` //商品图片1-中图
	ProductPic1Large  string `form:"product_pic1_large" json:"product_pic1_large,omitempty"`   //商品图片1-大图
	ProductPic2Small  string `form:"product_pic2_small" json:"product_pic2_small,omitempty"`   //商品图片2-小图
	ProductPic2Middle string `form:"product_pic2_middle" json:"product_pic2_middle,omitempty"` //商品图片2-中图
	ProductPic2Large  string `form:"product_pic2_large" json:"product_pic2_large,omitempty"`   //商品图片2-大图
	ProductPic3Small  string `form:"product_pic3_small" json:"product_pic3_small,omitempty"`   //商品图片3-小图
	ProductPic3Middle string `form:"product_pic3_middle" json:"product_pic3_middle,omitempty"` //商品图片3-中图
	ProductPic3Large  string `form:"product_pic3_large" json:"product_pic3_large,omitempty"`   //商品图片3-大图
	Thumbnail         string `form:"thumbnail" json:"thumbnail,omitempty"`                     //商品短视频缩略图
	ShortVideo        string `form:"short_video" json:"short_video,omitempty"`                 //商品短视频
	CreateAt          int64  `form:"create_at" json:"create_at,omitempty"`                     //创建时刻， 也就是上架时刻
	ModifyAt          int64  `form:"modify_at" json:"modify_at,omitempty"`                     //最后修改时间
}

package models

type OrderImagesOnBlockChainHistory struct {
	OrderID          string  `json:"order_id"`                              //订单ID
	ProductID        string  `form:"product_id" json:"product_id"`          //商品ID
	AttachHash       string  `form:"attach_hash" json:"attach_hash"`        //订单内容hash
	BuyUsername      string  `json:"buy_username" validate:"required"`      //买家注册号
	BusinessUsername string  `json:"business_username" validate:"required"` //商户注册号
	Cost             float64 `json:"cost" validate:"required"`              //本订单的总金额

	//订单图片在买家的oss objectID 暂时支持1张图片, 等迁移到Gorm2.0并重构数据库后改为数组
	BuyerOssImages string `json:"buyer_images" validate:"required"`

	//订单图片在商户的oss objectID
	BusinessOssImages string `json:"business_images" validate:"required"`
}

package models

type OrderImagesOnBlockChainHistory struct {
	OrderID          string  `json:"order_id"`                              //订单ID
	ProductID        string  `form:"product_id" json:"product_id"`          //商品ID
	AttachHash       string  `form:"attach_hash" json:"attach_hash"`        //订单内容hash
	BuyUsername      string  `json:"buy_username" validate:"required"`      //买家注册号
	BusinessUsername string  `json:"business_username" validate:"required"` //商户注册号
	Cost             float64 `json:"cost" validate:"required"`              //本订单的总金额

	//订单图片在商户的oss objectID
	BusinessOssImage string `json:"business_images" validate:"required"`
}

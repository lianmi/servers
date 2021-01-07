package models

import (
	"github.com/lianmi/servers/internal/pkg/models/global"
	"gorm.io/gorm"
	"time"
)

//redis里订单数据
type OrderInfo struct {
	OrderID          string  `json:"order_id"`                              //订单ID
	ProductID        string  ` json:"product_id"`                           //商品ID
	AttachHash       string  ` json:"attach_hash"`                          //订单内容哈希
	BuyerUsername    string  `json:"buyer_username" validate:"required"`    //买家注册号
	BusinessUsername string  `json:"business_username" validate:"required"` //商户注册号
	Cost             float64 `json:"cost" validate:"required"`              //本订单的总金额
	State            int     `json:"state"`                                 //订单类型
	IsPayed          bool    `json:"is_payed"`                              //此订单支付状态， true- 支付完成，false-未支付
	IsUrge           bool    `json:"is_urge"`                               //催单

}

/*
服务端的订单图片上链历史表
*/
type OrderImagesHistory struct {
	global.LMC_Model

	OrderID          string  `gorm:"primarykey" json:"order_id"`            //订单ID
	ProductID        string  ` json:"product_id"`                           //商品ID
	BuyUsername      string  `json:"buy_username" validate:"required"`      //买家注册号
	BusinessUsername string  `json:"business_username" validate:"required"` //商户注册号
	Cost             float64 `json:"cost" validate:"required"`              //本订单的总金额

	//订单图片在商户的oss objectID
	BusinessOssImage string `json:"business_image" validate:"required"`

	BlockNumber uint64 `json:"block_number"` //成功执行合约的所在区块高度
	TxHash      string `json:"tx_hash" `     //交易哈希

}

//BeforeCreate CreatedAt赋值
func (o *OrderImagesHistory) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (o *OrderImagesHistory) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

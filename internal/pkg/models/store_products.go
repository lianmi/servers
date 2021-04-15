package models

import (
	"gorm.io/gorm"
	"time"
)

// 存储商户的商品表
type StoreProductItems struct {
	UUID        string         `gorm:"primarykey;type:char(128)" form:"-" json:"-" ` //格式
	CreatedAt   time.Time      `json:"-"`
	UpdatedAt   time.Time      `json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	StoreUUID   string         `json:"store_uuid"`  // 店铺的uuid
	ProductId   string         `json:"product_id"`  // 支持的彩种id
	UserFormat  string         `json:"user_format"` // 用户定义的展示格式 自定义
	UserData    string         `json:"user_data"`   // 用户定义的json 这个商品的特殊定义 , 用户可以自定义描述 , 将以是json
	ProductInfo GeneralProduct `json:"-" gorm:"foreignKey:ProductId;references:ProductId"`
}

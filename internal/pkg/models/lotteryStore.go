package models

import (
	"github.com/lianmi/servers/internal/pkg/models/global"
)

type LotteryStore struct {
	global.LMC_Model
	Keyword   string `form:"keyword" json:"keyword,omitempty" `       //关键字 体彩 福彩
	MapID     string `gorm:"primarykey" form:"map_id" json:"map_id" ` //高德地图的id
	Province  string `form:"province" json:"province,omitempty" `     //省份, 如广东省
	City      string `form:"city" json:"city,omitempty" `             //城市，如广州市
	Area      string `form:"area" json:"area,omitempty" `             //区，如天河区
	Address   string `form:"address" json:"address,omitempty" `       //地址
	StoreName string `form:"store_name" json:"store_name,omitempty" ` //店铺名称
	Longitude string `form:"longitude" json:"longitude,omitempty" `   //商户地址的经度
	Latitude  string `form:"latitude" json:"latitude,omitempty" `     //商户地址的纬度
	Phones    string `form:"phones" json:"phones,omitempty" `         //联系手机或电话, 以半角逗号隔开
	Photos    string `form:"photos" json:"photos" `                   //店铺外景照片, 以半角逗号隔开
	StoreType int    `form:"store_type" json:"store_type"`            //店铺类型, 1-福彩 2-体彩
	Status    int    `form:"status" json:"status"`                    //状态，0-预，1-已提交
}

//用于查找
type LotteryStoreReq struct {
	Keyword  string `form:"keyword"`
	Province string `form:"province"`
	City     string `form:"city"`
	Area     string `form:"area"`
	Address  string `form:"address"` //模糊查找

	Limit  int `form:"limit"`
	Offset int `form:"offset"`
	// Status int `form:"status"`
}

package models

import (
	"time"

	"gorm.io/gorm"
)

//店铺表
//占位, 通过爬虫得到的各种类型的店铺，填充搜索页面，当用户点击进去后，页面会提醒用户此店铺尚未注册开通，引导真正的店铺注册
//另外，如果是彩票店，则有若干真实存在的店铺，用户点击后可以跳转到真正的店铺下单
//当此店铺的商户提交审核后，后台需要删除此记录

type Store struct {
	ID                uint64  `gorm:"primary_key" form:"id" json:"id,omitempty"`                             //自动递增id
	CreatedAt         int64   `form:"created_at" json:"created_at,omitempty"`                                //创建时刻,毫秒
	UpdatedAt         int64   `form:"updated_at" json:"updated_at,omitempty"`                                //更新时刻,毫秒
	StoreUUID         string  `form:"store_uuid" json:"store_uuid" `                                         //店铺的uuid
	StoreType         int     `form:"store_type" json:"store_type"`                                          //店铺类型,对应Global.proto里的StoreType枚举
	BusinessUsername  string  `form:"business_username" json:"business_username" `                           //商户注册号
	Introductory      string  `gorm:"type:longtext;null" form:"introductory" json:"introductory,omitempty" ` //商店简介 Text文本类型
	Province          string  `form:"province" json:"province,omitempty" `                                   //省份, 如广东省
	City              string  `form:"city" json:"city,omitempty" `                                           //城市，如广州市
	County            string  `form:"county" json:"county,omitempty" `                                       //区，如天河区
	Street            string  `form:"street" json:"street,omitempty" `                                       //街道
	Address           string  `form:"address" json:"address,omitempty" `                                     //地址
	Branchesname      string  `form:"branches_name" json:"branches_name,omitempty" `                         //网点名称
	LegalPerson       string  `form:"legal_person" json:"legal_person,omitempty" `                           //法人姓名
	LegalIdentityCard string  `form:"legal_identity_card" json:"legal_identity_card,omitempty" `             //法人身份证
	Longitude         float64 `form:"longitude" json:"longitude,omitempty" `                                 //商户地址的经度
	Latitude          float64 `form:"latitude" json:"latitude,omitempty" `                                   //商户地址的纬度
	WeChat            string  `form:"wechat" json:"wechat,omitempty" `                                       //商户联系人微信号
	Keys              string  `form:"keys" json:"keys,omitempty" `                                           //商户经营范围搜索关键字
	LicenseURL        string  `form:"license_url" json:"license_url" `                                       //商户营业执照阿里云url
	AuditState        int     `form:"audit_state" json:"audit_state"`                                        //审核状态，0-预审核，1-审核通过, 2-占位
}

//BeforeCreate CreatedAt赋值
func (l *Store) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (l *Store) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

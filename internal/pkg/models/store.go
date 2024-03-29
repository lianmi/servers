package models

import (
	"time"

	"github.com/lianmi/servers/internal/pkg/models/global"
	"gorm.io/gorm"
)

//店铺表
//占位, 通过爬虫得到的各种类型的店铺，填充搜索页面，当用户点击进去后，页面会提醒用户此店铺尚未注册开通，引导真正的店铺注册
//另外，如果是彩票店，则有若干真实存在的店铺，用户点击后可以跳转到真正的店铺下单
//当此店铺的商户提交审核后，后台需要删除此记录

type Store struct {
	global.LMC_Model

	StoreUUID             string `gorm:"primarykey" form:"store_uuid" json:"store_uuid" `             //店铺的uuid
	StoreType             int    `form:"store_type" json:"storeType"`                                 //店铺类型,对应Global.proto里的StoreType枚举
	ImageURL              string `form:"image_url" json:"imageUrl" `                                  //店铺外景照片或产品图片
	BusinessUsername      string `form:"business_username" json:"business_username" `                 //商户注册号
	Introductory          string `gorm:"type:longtext;null" form:"introductory" json:"introductory" ` //商店简介 Text文本类型
	Keys                  string `form:"keys" json:"keys" `                                           //商户经营范围搜索关键字
	ContactMobile         string `form:"contact_mobile" json:"contactMobile" `                        //联系电话
	WeChat                string `form:"wechat" json:"wechat" `                                       //商户联系人微信号
	Branchesname          string `form:"branches_name" json:"branches_name" `                         //网点名称
	OpeningHours          string `form:"opening_hours" json:"openingHours"`                           //营业时间
	Province              string `form:"province" json:"province" `                                   //省份, 如广东省
	City                  string `form:"city" json:"city" `                                           //城市，如广州市
	Area                  string `form:"area" json:"area" `                                           //区，如天河区
	// Street                string `form:"street" json:"street" `                                       //街道
	Address               string `form:"address" json:"address" `                                     //地址
	LegalPerson           string `form:"legal_person" json:"legalPerson" `                            //法人姓名
	LegalIdentityCard     string `form:"legal_identity_card" json:"legalIdentityCard" `               //法人身份证
	Longitude             string `form:"longitude" json:"longitude" `                                 //商户地址的经度
	Latitude              string `form:"latitude" json:"latitude" `                                   //商户地址的纬度
	LicenseURL            string `form:"license_url" json:"businessLicenseUrl" `                      //商户营业执照阿里云url
	AuditState            int    `form:"audit_state" json:"audit_state"`                              //审核状态，0-预审核，1-审核通过, 2-占位
	DefaultOPK            string `form:"default_opk" json:"default_opk"`                              //商户的默认OPK
	BusinessCode          string `form:"business_code" json:"businessCode"`                           //商户的网点编码，适合彩票店或连锁网点
	NotaryServiceUsername string `form:"notary_service_username" json:"notary_service_username"`      //商户对应的公证处注册id
	IdCardFrontPhoto string `json:"idCardFrontPhoto"` // 身份证 正面
	IdCardBackPhoto  string `json:"idCardBackPhoto"`  // 身份证 背面
	CardOwner        string `json:"cardOwner"`        // 银行卡持有者姓名
	BankName         string `json:"bankName"`         // 银行名称
	BankBranch       string `json:"bankBranch"`       // 银行支行
	CardNumber       string `json:"cardNumber"`       // 银行卡号
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

//see: https://www.jianshu.com/p/a43c6d2f8bfb
//商店的点赞明细
/*
参数校验
对传入的参数进行null值判断
逻辑校验
对于用户点赞，用户不能重复点赞相同的商店
对于取消点赞，用户不能取消未点赞的商店
存入Redis
存入的数据主要有所有商店的点赞数、某商店的点赞数、用户点赞的商店
定时任务
通过定时【1小时执行一次】，从Redis读取数据持久化到MySQL中
*/
type StoreLike struct {
	global.LMC_Model

	BusinessUsername string `gorm:"primarykey" form:"business_username" json:"business_username" ` //商户注册号
	Username         string `form:"username" json:"username" `                                     //普通用户注册号
}

//用户点赞的店铺记录表
//每个用户只能点赞一次，如果是点赞过了，可以取消点赞
type UserLike struct {
	global.LMC_Model

	Username         string `gorm:"primarykey" form:"username" json:"username" ` //普通用户注册号
	BusinessUsername string `form:"business_username" json:"business_username" ` //商户注册号
}

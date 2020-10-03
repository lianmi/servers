package models

import (
	"time"

	"github.com/jinzhu/gorm"
	// pb "github.com/lianmi/servers/api/proto/user"
)

/*
此表是用于保存用户钱包地址及连米币最新余额及Eth余额
*/

type UserWallet struct {
	ID              uint64 `gorm:"primary_key" form:"id" json:"id,omitempty"` //自动递增id
	CreatedAt       int64  `form:"created_at" json:"created_at,omitempty"`    //创建时刻,毫秒
	UpdatedAt       int64  `form:"updated_at" json:"updated_at,omitempty"`    //更新时刻,毫秒
	Username        string `json:"username" `                                 //用户注册号，自动生成，字母 + 数字
	WalletAddress   string `json:"wallet_address" validate:"required"`        //用户链上地址，默认是用户HD钱包的第0号索引，用于存储Eth及连米币
	AmountETHString string `json:"amount_eth_string" validate:"required"`     //用户eth数量 wei单位, 由于是大数，所以用字符串类型代替
	AmountLNMC      int64  `json:"amount_lnmc" validate:"required"`           //用户连米币数量
}

//BeforeCreate CreatedAt赋值
func (w *UserWallet) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (w *UserWallet) BeforeUpdate(scope *gorm.Scope) error {
	scope.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//此表是用于保存用户充值记录
//TODO
type LnmcDepositHistory struct {
	ID               uint64 `gorm:"primary_key" form:"id" json:"id,omitempty"` //自动递增id
	CreatedAt        int64  `form:"created_at" json:"created_at,omitempty"`    //创建时刻,毫秒
	UpdatedAt        int64  `form:"updated_at" json:"updated_at,omitempty"`    //更新时刻,毫秒
	Username         string `json:"username" `                                 //用户注册号，自动生成，字母 + 数字
	WalletAddress    string `json:"wallet_address" validate:"required"`        //用户链上地址，默认是用户HD钱包的第0号索引，用于存储Eth及连米币
	AmountLNMCBefore int64  `json:"amount_lnmc_before" validate:"required"`    //充值前用户连米币数量
	DepositAmount    int64  `json:"deposit_amount" validate:"required"`        //充值金额，单位是人民币
	PaymentType      int    `json:"payment_type" validate:"required"`          //第三方支付方式 1- 支付宝， 2-微信 3-银行卡
	AmountLNMCAfter  int64  `json:"amount_lnmc_after" validate:"required"`     //充值后用户连米币数量
	BlockNumber      int64  `json:"block_number" validate:"required"`          //交易成功打包的区块高度
	TxHash           string `json:"tx_hash" validate:"required"`               //交易成功打包的区块哈希
}

//BeforeCreate CreatedAt赋值
func (w *LnmcDepositHistory) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (w *LnmcDepositHistory) BeforeUpdate(scope *gorm.Scope) error {
	scope.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//此表是用于保存用户LNMC连米币转账及支付记录
//TODO
type LnmcTransferHistory struct {
	ID          uint64 `gorm:"primary_key" form:"id" json:"id,omitempty"` //自动递增id
	CreatedAt   int64  `form:"created_at" json:"created_at,omitempty"`    //创建时刻,毫秒
	UpdatedAt   int64  `form:"updated_at" json:"updated_at,omitempty"`    //更新时刻,毫秒
	Username    string `json:"username" `                                 //用户注册号，自动生成，字母 + 数字
	AccountETH  string `json:"account_eth" validate:"required"`           //用户链上地址，默认是用户HD钱包的第0号索引，用于存储Eth
	AccountLNMC string `json:"account_lnmc" validate:"required"`          //用户链上地址，默认是用户HD钱包的第1号索引，用于存储连米币
	AmountETH   int64  `json:"amount_eth" validate:"required"`            //用户eth数量
	AmountLNMC  int64  `json:"amount_lnmc" validate:"required"`           //用户连米币数量
}

//此表是用于保存用户ETH转账及支付记录
//TODO
type ETHTransferHistory struct {
	ID          uint64 `gorm:"primary_key" form:"id" json:"id,omitempty"` //自动递增id
	CreatedAt   int64  `form:"created_at" json:"created_at,omitempty"`    //创建时刻,毫秒
	UpdatedAt   int64  `form:"updated_at" json:"updated_at,omitempty"`    //更新时刻,毫秒
	Username    string `json:"username" `                                 //用户注册号，自动生成，字母 + 数字
	AccountETH  string `json:"account_eth" validate:"required"`           //用户链上地址，默认是用户HD钱包的第0号索引，用于存储Eth
	AccountLNMC string `json:"account_lnmc" validate:"required"`          //用户链上地址，默认是用户HD钱包的第1号索引，用于存储连米币
	AmountETH   int64  `json:"amount_eth" validate:"required"`            //用户eth数量
	AmountLNMC  int64  `json:"amount_lnmc" validate:"required"`           //用户连米币数量
}

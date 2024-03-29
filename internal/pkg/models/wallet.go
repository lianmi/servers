package models

import (
	// "encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lianmi/servers/internal/pkg/models/global"
	"gorm.io/gorm"
	"math/big"
	"time"
)

/*
此表是用于保存用户钱包地址及连米币最新余额及Eth余额
*/

type UserWallet struct {
	global.LMC_Model

	Username string `gorm:"primarykey"  json:"username" validate:"required"` //用户注册号， 对应User表的username字段

	WalletAddress   string `json:"wallet_address" validate:"required"`    //用户链上地址，默认是用户HD钱包的第0号索引，用于存储Eth及连米币
	AmountETHString string `json:"amount_eth_string" validate:"required"` //用户eth数量 wei单位, 由于是大数，所以用字符串类型代替
	AmountLNMC      int64  `json:"amount_lnmc" validate:"required"`       //用户连米币数量
}

//BeforeCreate CreatedAt赋值
func (w *UserWallet) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (w *UserWallet) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//此表是用于保存用户充值记录
type LnmcDepositHistory struct {
	global.LMC_Model

	Username          string  `json:"username" validate:"required"`           //用户注册号
	WalletAddress     string  `json:"wallet_address" validate:"required"`     //用户链上地址，默认是用户HD钱包的第0号索引，用于存储Eth及连米币
	BalanceLNMCBefore int64   `json:"amount_lnmc_before" validate:"required"` //充值前用户连米币数量
	RechargeAmount    float64 `json:"recharge_amount" validate:"required"`    //充值金额，单位是人民币
	PaymentType       int     `json:"payment_type" validate:"required"`       //第三方支付方式 1- 支付宝， 2-微信 3-银行卡
	BalanceLNMCAfter  int64   `json:"amount_lnmc_after" validate:"required"`  //充值后用户连米币数量
	BlockNumber       uint64  `json:"block_number" validate:"required"`       //交易成功打包的区块高度
	TxHash            string  `json:"tx_hash" validate:"required"`            //交易成功打包的区块哈希
}

//BeforeCreate CreatedAt赋值
func (w *LnmcDepositHistory) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (w *LnmcDepositHistory) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//此表是用于保存用户LNMC连米币转账及支付记录
type LnmcTransferHistory struct {
	global.LMC_Model

	UUID string `gorm:"primarykey" form:"uuid" json:"uuid,omitempty"` //uuid

	Username            string `json:"username" validate:"required"`              //发送方用户注册号
	ToUsername          string `json:"to_username" validate:"required"`           // 接收方注册号
	WalletAddress       string `json:"wallet_address" validate:"required"`        //发送方用户链上地址，默认是用户HD钱包的第0号索引，用于存储连米币
	ToWalletAddress     string `json:"to_wallet_address" validate:"required"`     //接收方用户链上地址，默认是用户HD钱包的第0号索引，用于存储连米币
	BalanceLNMCBefore   uint64 `json:"amount_lnmc_before" validate:"required"`    //发送方用户在转账时刻的连米币数量
	AmountLNMC          uint64 `json:"amount_lnmc" validate:"required"`           //本次转账的用户连米币数量
	Content             string `json:"content"`                                   //附言
	BalanceLNMCAfter    uint64 `json:"amount_lnmc_after" validate:"required"`     //发送方用户在转账之后的连米币数量
	Bip32Index          uint64 `json:"bip32_index" validate:"required"`           //平台HD钱包Bip32派生索引号
	ContractBlockNumber uint64 `json:"contract_block_number" validate:"required"` //多签合约所在区块高度
	ContractHash        string `json:"contract_hash" validate:"required"`         //多签合约的哈希
	State               int    `json:"state" validate:"required"`                 //多签合约执行状态，0-默认未执行，1-已执行
	OrderID             string `json:"order_id"`                                  //如果非空，则此次支付是对订单的支付，如果空，则为普通转账
	BlockNumber         uint64 `json:"block_number"`                              //成功执行合约的所在区块高度
	TxHash              string `json:"tx_hash" `                                  //交易哈希
}

//BeforeCreate CreatedAt赋值
func (w *LnmcTransferHistory) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (w *LnmcTransferHistory) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//此表是用于保存用户LNMC连米币提现 记录
type LnmcWithdrawHistory struct {
	global.LMC_Model

	WithdrawUUID      string `json:"withdraw_uuid" validate:"required"`      //提现单编号，UUID
	Username          string `json:"username" validate:"required"`           //发送方用户注册号
	Bank              string `json:"bank" validate:"required"`               //银行名称
	BankCard          string `json:"bank_card" validate:"required"`          //银行卡号
	CardOwner         string `json:"card_owner" validate:"required"`         //银行卡持有人
	WalletAddress     string `json:"wallet_address" validate:"required"`     //发送方用户链上地址，默认是用户HD钱包的第0号索引，用于存储连米币
	BalanceLNMCBefore uint64 `json:"amount_lnmc_before" validate:"required"` //发送方用户在转账时刻的连米币数量
	AmountLNMC        uint64 `json:"amount_lnmc" validate:"required"`        //本次提现的用户连米币数量
	BalanceLNMCAfter  uint64 `json:"amount_lnmc_after" validate:"required"`  //本次提现之后的用户连米币数量
	State             int    `json:"state" validate:"required"`              //提现进度状态，0-默认未执行，1-已执行
	BlockNumber       uint64 `json:"block_number"`                           //成功执行提现的所在区块高度
	TxHash            string `json:"tx_hash" `                               //交易哈希
	Fee               uint64 `json:"fee" validate:"required"`                //本次提现的佣金总额
}

//BeforeCreate CreatedAt赋值
func (w *LnmcWithdrawHistory) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (w *LnmcWithdrawHistory) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//此表是用于保存用户收款记录，收款有两个来源，一是订单支付，二是转账

type LnmcCollectionHistory struct {
	global.LMC_Model

	FromUsername      string `json:"from_username" validate:"required"`       //发送方用户注册号
	FromWalletAddress string `json:"from_wallet_address" validate:"required"` //发送方用户链上地址，默认是用户HD钱包的第0号索引，用于存储连米币
	ToUsername        string `json:"to_username" validate:"required"`         // 接收方用户注册号
	ToWalletAddress   string `json:"to_wallet_address" validate:"required"`   //接收方用户链上地址，默认是用户HD钱包的第0号索引，用于存储连米币
	BalanceLNMCBefore uint64 `json:"amount_lnmc_before" validate:"required"`  //发送方用户在转账时刻的连米币数量
	AmountLNMC        uint64 `json:"amount_lnmc" validate:"required"`         //本次转账的用户连米币数量
	BalanceLNMCAfter  uint64 `json:"amount_lnmc_after" validate:"required"`   //发送方用户在转账之后的连米币数量
	Bip32Index        uint64 `json:"bip32_index" validate:"required"`         //平台HD钱包Bip32派生索引号
	OrderID           string `json:"order_id"`                                //如果非空，则此次支付是对订单的支付，如果空，则为普通转账
	BlockNumber       uint64 `json:"block_number"`                            //成功执行合约的所在区块高度
	TxHash            string `json:"tx_hash" `                                //交易哈希
}

//BeforeCreate CreatedAt赋值
func (w *LnmcCollectionHistory) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (w *LnmcCollectionHistory) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//裸交易结构体
type RawDesc struct {
	//接收者的钱包地址
	DestinationAddress string `protobuf:"bytes,6,opt,name=destination_address,proto3" json:"destination_address,omitempty"`
	//nonce
	Nonce uint64 `protobuf:"fixed64,1,opt,name=nonce,proto3" json:"nonce,omitempty"`
	// gas价格
	GasPrice uint64 `protobuf:"fixed64,2,opt,name=gasPrice,proto3" json:"gasPrice,omitempty"`
	// 最低gas
	GasLimit uint64 `protobuf:"fixed64,3,opt,name=gasLimit,proto3" json:"gasLimit,omitempty"`
	//链id
	ChainID uint64 `protobuf:"fixed64,4,opt,name=chainID,proto3" json:"chainID,omitempty"`
	// 交易数据
	Txdata []byte `protobuf:"bytes,5,opt,name=txdata,proto3" json:"txdata,omitempty"`

	//要转账的代币数量
	Value uint64 `protobuf:"fixed64,7,opt,name=value,proto3" json:"value,omitempty"`
	//交易哈希
	TxHash string `protobuf:"bytes,6,opt,name=txHash,proto3" json:"txHash,omitempty"`

	//发币合约智能地址
	ContractAddress string `protobuf:"bytes,6,opt,name=contract_address,proto3" json:"contract_address,omitempty"`
}

//裸成功交易结构体
type HashInfo struct {
	IsPending   bool   `json:"is_pending"`
	BlockNumber uint64 `json:"block_number"` //成功执行合约的所在区块高度
	TxHash      string `json:"tx_hash" `     //交易哈希
	//nonce
	Nonce uint64 `protobuf:"fixed64,1,opt,name=nonce,proto3" json:"nonce,omitempty"`
	Gas   uint64 `protobuf:"fixed64,2,opt,name=gasPrice,proto3" json:"gasPrice,omitempty"`
	// 数据
	Data string `protobuf:"bytes,5,opt,name=data,proto3" json:"data,omitempty"`

	To string `protobuf:"bytes,6,opt,name=to,proto3" json:"to,omitempty"`
}

//此表是用于保存订单完成后的到账或撤单退款记录
type LnmcOrderTransferHistory struct {
	global.LMC_Model

	OrderID                   string  `json:"order_id"  validate:"required"`                    //订单ID
	PayType                   int     `json:"pay_type"  validate:"required"`                    //转账类型，1-订单完成，被买家确认后转账给商户，2-撤单及拒单退款
	ProductID                 string  `json:"product_id"  validate:"required"`                  //商品ID
	BuyUser                   string  `json:"buy_user" validate:"required"`                     //买家注册号
	BusinessUser              string  `json:"business_user" validate:"required"`                //商户注册号
	BuyUserWalletAddress      string  `json:"buy_user_wallet_address" validate:"required"`      //买家链上地址，默认是用户HD钱包的第0号索引，用于存储连米币
	BusinessUserWalletAddress string  `json:"business_user_wallet_address" validate:"required"` //商户链上地址，默认是用户HD钱包的第0号索引，用于存储连米币
	AttachHash                string  `json:"attach_hash" validate:"required"`                  //订单内容哈希，上链
	Bip32Index                uint64  `json:"bip32_index" validate:"required"`                  //买家对应平台HD钱包Bip32派生索引号
	BalanceLNMCBefore         uint64  `json:"amount_lnmc_before" validate:"required"`           //平台HD钱包在转账时刻的连米币数量
	OrderTotalAmount          float64 `json:"order_total_amount" validate:"required"`           //人民币格式的订单总金额
	AmountLNMC                uint64  `json:"amount_lnmc" validate:"required"`                  //本次转账的连米币数量, 无小数点
	BalanceLNMCAfter          uint64  `json:"amount_lnmc_after" validate:"required"`            //平台HD钱包在转账之后的连米币数量
	BlockNumber               uint64  `json:"block_number"`                                     //成功执行合约的所在区块高度
	TxHash                    string  `json:"tx_hash" `                                         //交易哈希
}

//BeforeCreate CreatedAt赋值
func (w *LnmcOrderTransferHistory) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (w *LnmcOrderTransferHistory) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//支付宝支付历史表
type AliPayHistory struct {
	global.LMC_Model

	TradeNo     string  `json:trade_no"  validate:"required"`
	Username    string  `json:"username"  validate:"required"`
	Subject     string  `json:"subject" validate:"required"`
	ProductCode string  `json:"product_code" validate:"required"`
	TotalAmount float64 `json:"total_amount" validate:"required"`
	Fee         float64 `json:"fee" validate:"required"` //本次充值的手续费
	IsPayed     bool    `json:"is_payed" validate:"required"`
}

//BeforeCreate CreatedAt赋值
func (w *AliPayHistory) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (w *AliPayHistory) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//系统服务费商品ID表
type SystemCharge struct {
	global.LMC_Model

	ChargeProductID string `json:"charge_product_id"  validate:"required"` //服务费的商品D
}

//BeforeCreate CreatedAt赋值
func (s *SystemCharge) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (s *SystemCharge) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//服务费收款历史, 在common.go里定义了接收服务费的账号，暂定为id3
type ChargeHistory struct {
	global.LMC_Model

	BuyerUsername    string  `json:"buyer_username"  validate:"required"`                    //买家
	BusinessUsername string  `json:"business_username"  validate:"required"`                 //商户
	ChargeProductID  string  `json:"charge_product_id"  validate:"required"`                 //服务费的商品D
	ChargeOrderID    string  `gorm:"primarykey" json:"charge_order_id"  validate:"required"` //本次服务费的订单ID
	BusinessOrderID  string  `json:"business_order_id"  validate:"required"`                 //商品订单ID, 买家支付的订单ID
	OrderTotalAmount float64 `json:"order_total_amount" validate:"required"`                 //人民币格式的订单总金额
	IsVip            bool    `json:"is_vip" validate:"required"`                             //是否是Vip用户
	Rate             float64 `json:"rate" validate:"required"`                               //费率
	ChargeAmount     float64 `json:"charge_amount" validate:"required"`                      //服务费
	BlockNumber      uint64  `json:"block_number"`                                           //所在区块高度
	TxHash           string  `json:"tx_hash" `                                               //交易哈希
	IsPayed          bool    `json:"is_payed" `                                              //是否已经支付
}

//BeforeCreate CreatedAt赋值
func (s *ChargeHistory) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (s *ChargeHistory) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//链上交易打包成功后返回的交易数据
type TxInfo struct {
	BlockNumber uint64
	HashHex     string
	Value       string
	Gas         uint64
	GasPrice    uint64
	GasCost     uint64
	EthAmount   float64
	Nonce       uint64
	Data        []byte
	From        string
	Status      uint64
}

// 根据erc20智能合约地址, 获取所需的块范围交易日志数据 LogTransfer ..
type LogTransfer struct {
	From   common.Address
	To     common.Address
	Tokens *big.Int
}

// LogApproval ..
type LogApproval struct {
	TokenOwner common.Address
	Spender    common.Address
	Tokens     *big.Int
}

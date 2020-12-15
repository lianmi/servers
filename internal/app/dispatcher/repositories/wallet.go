package repositories

import (
	"fmt"
	"strconv"

	// "github.com/gomodule/redigo/redis"
	Wallet "github.com/lianmi/servers/api/proto/wallet"
	LMCommon "github.com/lianmi/servers/internal/common"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"

	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/smartwalle/alipay/v3"
	"github.com/smartwalle/xid"
)

func (s *MysqlLianmiRepository) PreAlipay(username, totalAmount string) (*Wallet.PreAlipayRsp, error) {
	var err error
	var aliClient *alipay.Client

	var tradeNo = fmt.Sprintf("%d", xid.Next())

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	// 第三个参数是沙箱(false) , 正式环境是 true
	if aliClient, err = alipay.New(LMCommon.AlipayAppId, LMCommon.AppPrivateKey, true); err != nil {
		s.logger.Error("初始化支付宝失败", zap.Error(err))
		return nil, err
	}

	//使用支付宝公钥, 只能二选一 , 所以我选了支付宝公钥
	if err = aliClient.LoadAliPayPublicKey(LMCommon.AlipayPublicKey); err != nil {
		s.logger.Error("加载支付宝公钥发生错误", zap.Error(err))
		return nil, err
	} else {
		s.logger.Debug("加载支付宝公钥成功")
	}

	var productCode = "deposit_" + totalAmount
	var p = alipay.TradeAppPay{}
	p.NotifyURL = LMCommon.ServerDomain + "/v1/wallet/alipay/notify"
	p.ReturnURL = LMCommon.ServerDomain + "/v1/wallet/alipay/callback"
	p.Body = username //body保存用户的注册账号
	p.Subject = "支付充值:" + tradeNo + "_" + totalAmount
	p.OutTradeNo = tradeNo
	p.TotalAmount = totalAmount
	p.ProductCode = productCode

	param, err := aliClient.TradeAppPay(p)
	if err != nil {
		s.logger.Error("TradeAppPay发生错误", zap.Error(err))
	}
	s.logger.Debug("TradeAppPay param", zap.String("param", param))

	//将订单号保存到redis里，以便支付宝服务器回调后查找出支付内容
	preAlipayKey := fmt.Sprintf("PreAlipay:%s", tradeNo)

	_, err = redisConn.Do("HMSET",
		preAlipayKey,
		"Username", username,
		"Subject", "支付充值:"+tradeNo+"_"+totalAmount,
		"TotalAmount", totalAmount,
		"ProductCode", productCode,
		"IsPayed", false,
	)

	amount, err := strconv.ParseFloat(totalAmount, 64)
	if err != nil {
		s.logger.Error("增加ParseFloat失败", zap.Error(err))
		return nil, err
	}
	//保存到MySQL AliPayHistory表
	aliPayHistory := &models.AliPayHistory{
		TradeNo:     tradeNo,
		Username:    username,
		Subject:     "支付充值:" + tradeNo + "_" + totalAmount,
		ProductCode: productCode,
		TotalAmount: amount,
		Fee:         amount * 0.06,
		IsPayed:     false,
	}
	if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(aliPayHistory).Error; err != nil {
		s.logger.Error("增加AliPayHistory表失败", zap.Error(err))
		return nil, err
	} else {
		s.logger.Debug("增加AliPayHistory表成功")
	}
	return &Wallet.PreAlipayRsp{
		TradeNo:    tradeNo,
		Signedinfo: param,
	}, nil

}

//支付宝付款成功
func (s *MysqlLianmiRepository) AlipayDone(outTradeNo string) error {
	// var err error
	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	preAlipayKey := fmt.Sprintf("PreAlipay:%s", outTradeNo)
	redisConn.Do("HSET",
		preAlipayKey,
		"Status", true,
	)
	result := s.db.Model(&models.AliPayHistory{}).Where(&models.AliPayHistory{
		TradeNo: outTradeNo,
	}).Update("is_payed", true) //将Status变为true
	if result.Error != nil {
		s.logger.Error("封号失败", zap.Error(result.Error))
		return result.Error
	}

	//TODO, 增加用户的钱包余额

	return nil

}

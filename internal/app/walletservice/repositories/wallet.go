package repositories

import (
	"context"
	"fmt"

	"github.com/gomodule/redigo/redis"

	// Global "github.com/lianmi/servers/api/proto/global"
	"net/url"
	"strconv"

	Wallet "github.com/lianmi/servers/api/proto/wallet"
	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/lianmi/servers/internal/pkg/wxpay"
	"github.com/pkg/errors"
	"github.com/smartwalle/alipay/v3"
	"github.com/smartwalle/xid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	// uuid "github.com/satori/go.uuid"
	"time"

	"github.com/lianmi/servers/util/dateutil"
)

type Amount struct{ Total float64 }

type WalletRepository interface {
	DoPreAlipay(ctx context.Context, req *Wallet.PreAlipayReq) (*Wallet.PreAlipayResp, error)

	DoPreWXpay(ctx context.Context, req *Wallet.PreWXpayReq) (*Wallet.PreWXpayResp, error)

	SaveDepositForPay(tradeNo, hash string, blockNumber, balanceLNMC uint64) error

	AddLnmcOrderTransferHistory(lnmcOrderTransferHistory *models.LnmcOrderTransferHistory) error

	AddUserWallet(username, walletAddress, amountETHString string) (err error)

	//增加用户充值历史记录
	AddDepositHistory(lnmcDepositHistory *models.LnmcDepositHistory) (err error)

	//增加预审核转账历史记录
	AddLnmcTransferHistory(lmnccTransferHistory *models.LnmcTransferHistory) (err error)

	//修改转账历史记录
	UpdateLnmcTransferHistory(lmncTransferHistory *models.LnmcTransferHistory) (err error)

	//增加预审核提现历史记录
	AddLnmcWithdrawHistory(lnmcWithdrawHistory *models.LnmcWithdrawHistory) (err error)

	//修改提现历史记录
	UpdateLnmcWithdrawHistory(lnmcWithdrawHistory *models.LnmcWithdrawHistory) (err error)

	AddeCollectionHistory(lnmcCollectionHistory *models.LnmcCollectionHistory) (err error)

	GetPages(model interface{}, out interface{}, pageIndex, pageSize int, totalCount *int64, where interface{}, orders ...string) error

	GetCollectionHistorys(toUsername, fromUsername string, startAt, endAt uint64, pageNum int, pageSize int, total *int64) ([]*models.LnmcCollectionHistory, error)


	GetDepositHistorys(username string, PageNum int, PageSize int, total *int64, where interface{}) []*models.LnmcDepositHistory

	GetWithdrawHistorys(username string, PageNum int, PageSize int, total *int64, where interface{}) []*models.LnmcWithdrawHistory

	GetTransferHistorys(username string, PageNum int, PageSize int, total *int64, where interface{}) []*models.LnmcTransferHistory

	GetDepositInfo(txHash string) (*models.LnmcDepositHistory, error)

	GetWithdrawInfo(txHash string) (*models.LnmcWithdrawHistory, error)

	GetTransferInfo(txHash string) (*models.LnmcTransferHistory, error)

	//根据PayType获取到VIP价格
	GetVipUserPrice(payType int) (*models.VipPrice, error)

	//根据productID获取到VIP价格
	GetVipUserPriceByProductID(productID string) (*models.VipPrice, error)

	//会员付费成功后，按系统设定的比例进行佣金计算及写库， 需要新增3条佣金amount记录
	AddCommission(orderTotalAmount float64, username, orderID string) error

	//购买Vip后，增加用户的到期时间
	AddVipEndDate(username string, endTime int64) error

	// 修改ChargeHistory的支付及哈希区块
	UpdateChargeHistoryForPayed(chargeHistory *models.ChargeHistory) error
}

type MysqlWalletRepository struct {
	logger    *zap.Logger
	db        *gorm.DB
	redisPool *redis.Pool
	base      *BaseRepository
}

func NewMysqlWalletRepository(logger *zap.Logger, db *gorm.DB, redisPool *redis.Pool) WalletRepository {
	return &MysqlWalletRepository{
		logger:    logger.With(zap.String("type", "WalletRepository")),
		db:        db,
		redisPool: redisPool,
		base:      NewBaseRepository(logger, db),
	}
}

//调用支付宝SDK生成签名支付信息数据
func (m *MysqlWalletRepository) DoPreAlipay(ctx context.Context, req *Wallet.PreAlipayReq) (*Wallet.PreAlipayResp, error) {
	var err error
	var aliClient *alipay.Client

	var tradeNo = fmt.Sprintf("%d", xid.Next())

	redisConn := m.redisPool.Get()
	defer redisConn.Close()

	// 第三个参数是沙箱(false) , 正式环境是 true
	if aliClient, err = alipay.New(LMCommon.AlipayAppId, LMCommon.AppPrivateKey, true); err != nil {
		m.logger.Error("初始化支付宝失败", zap.Error(err))
		return nil, err
	}

	//使用支付宝公钥, 只能二选一 , 所以我选了支付宝公钥
	if err = aliClient.LoadAliPayPublicKey(LMCommon.AlipayPublicKey); err != nil {
		m.logger.Error("加载支付宝公钥发生错误", zap.Error(err))
		return nil, err
	} else {
		m.logger.Debug("加载支付宝公钥成功")
	}

	var productCode = "deposit_" + fmt.Sprintf("%f", req.TotalAmount)
	var subject = "支付充值:" + tradeNo + "_" + fmt.Sprintf("%f", req.TotalAmount)
	var p = alipay.TradeAppPay{}
	p.NotifyURL = LMCommon.ServerDomain + "/v1/wallet/alipay/notify"
	p.ReturnURL = LMCommon.ServerDomain + "/v1/wallet/alipay/callback"
	p.Body = req.Username //body保存用户的注册账号
	p.Subject = subject
	p.OutTradeNo = tradeNo
	p.TotalAmount = fmt.Sprintf("%f", req.TotalAmount)
	p.ProductCode = productCode

	param, err := aliClient.TradeAppPay(p)
	if err != nil {
		m.logger.Error("TradeAppPay发生错误", zap.Error(err))
		return nil, err
	}
	m.logger.Debug("TradeAppPay param", zap.String("param", param))

	//将订单号保存到redis里，以便支付宝服务器回调后查找出支付内容
	preAlipayKey := fmt.Sprintf("PreAlipay:%s", tradeNo)

	_, err = redisConn.Do("HMSET",
		preAlipayKey,
		"Username", req.Username,
		"Subject", subject,
		"TotalAmount", req.TotalAmount,
		"ProductCode", productCode,
		"IsPayed", false,
	)

	//保存到MySQL AliPayHistory表
	aliPayHistory := &models.AliPayHistory{
		TradeNo:     tradeNo,
		Username:    req.Username,
		Subject:     subject,
		ProductCode: productCode,
		TotalAmount: req.TotalAmount,
		Fee:         req.TotalAmount * 0.06,
		IsPayed:     false,
	}
	if err := m.db.Clauses(clause.OnConflict{DoNothing: true}).Create(aliPayHistory).Error; err != nil {
		m.logger.Error("增加AliPayHistory表失败", zap.Error(err))
		return nil, err
	} else {
		m.logger.Debug("增加AliPayHistory表成功")
	}
	return &Wallet.PreAlipayResp{
		TradeNo:    tradeNo,
		Signedinfo: param,
	}, nil

}

//向微信官方支付服务器发起预支付
func (m *MysqlWalletRepository) DoPreWXpay(ctx context.Context, req *Wallet.PreWXpayReq) (*Wallet.PreWXpayResp, error) {
	var err error
	var client *wxpay.Client
	client = wxpay.New(LMCommon.WXAppID, LMCommon.WXApiKey, LMCommon.WXMchID, true)
	if client == nil {
		m.logger.Error("wxpay.Client 无法创建")
		return nil, errors.Wrap(err, "wx  client cannot create!")
	}
	err = client.LoadCertFromBase64(LMCommon.WXCertDateBase64) //装载本地证书
	if err != nil {
		m.logger.Error("LoadCertFromBase64 错误 ", zap.Error(err))
		return nil, err
	}

	var p = wxpay.UnifiedOrderParam{}
	p.Body = req.Body //充值测试
	p.NotifyURL = LMCommon.ServerDomain + "/v1/wallet/wxpaynotify"
	p.TradeType = wxpay.TradeTypeApp //app支付
	p.SpbillCreateIP = req.ClientIP
	p.TotalFee = int(req.TotalAmount * 100)                          // 单位1分钱
	p.OutTradeNo = "" + strconv.FormatInt(time.Now().UnixNano(), 10) // 后面增加渠道编号

	result, err2 := client.UnifiedOrder(p)
	if err2 != nil {
		m.logger.Error("微信服务器返回错误", zap.Error(err2))

		return nil, err2
	}

	// 下面获取一下微信给的信息，然后组成串，签名发给客户端
	var wxmap = make(url.Values)
	wxmap.Set("appid", LMCommon.WXAppID)
	wxmap.Set("partnerid", LMCommon.WXMchID)
	wxmap.Set("prepayid", result.PrepayId)
	wxmap.Set("noncestr", result.NonceStr)
	timeStamp := strconv.FormatInt(time.Now().Unix(), 10)
	wxmap.Set("timestamp", timeStamp)
	wxmap.Set("package", "Sign=WXPay")
	var sign = wxpay.SignMD5(wxmap, LMCommon.WXApiKey)
	wxmap.Set("sign", sign)

	var re map[string]string = make(map[string]string, 0)
	for k, v := range wxmap {
		re[k] = v[0]
	}
	// RespData(c, http.StatusOK, 200, re)
	//TODO

	resp := &Wallet.PreWXpayResp{
		Appid:     LMCommon.WXAppID, //APPID
		Partnerid: LMCommon.WXMchID, //商户号
		Prepayid:  result.PrepayId,  //预支付id
		Noncestr:  result.NonceStr,  //nonce
		Package:   "Sign=WXPay",     //package
		Sign:      sign,             //签名
		Timestamp: timeStamp,        //时间戳
	}
	m.logger.Debug("微信支付的七个数据",
		zap.String("Appid", re["appid"]),
		zap.String("Partnerid", re["partnerid"]),
		zap.String("Prepayid", re["prepayid"]),
		zap.String("Noncestr", re["noncestr"]),
		zap.String("Package", re["package"]),
		zap.String("Sign", re["sign"]),
		zap.String("Timestamp", re["timeStamp"]),
	)
	return resp, nil
}

func (m *MysqlWalletRepository) SaveDepositForPay(tradeNo, hash string, blockNumber, balanceLNMC uint64) error {
	var err error
	var username string
	var walletAddress string
	var totalAmount float64

	redisConn := m.redisPool.Get()
	defer redisConn.Close()

	preAlipayKey := fmt.Sprintf("PreAlipay:%s", tradeNo)

	//获取username
	username, err = redis.String(redisConn.Do("HGET", preAlipayKey, "Username"))

	//获取充值金额
	totalAmount, err = redis.Float64(redisConn.Do("HGET", preAlipayKey, "TotalAmount"))

	result := m.db.Model(&models.AliPayHistory{}).Where(&models.AliPayHistory{
		TradeNo: tradeNo,
	}).Update("is_payed", true) //将Status变为true
	if result.Error != nil {
		m.logger.Error("将Status变为已支付", zap.Error(result.Error))
		return result.Error
	}

	walletAddress, err = redis.String(redisConn.Do("HGET", fmt.Sprintf("userWallet:%s", username), "WalletAddress"))
	if err != nil {
		m.logger.Error("HGET失败", zap.Error(result.Error))
		return err
	}

	//保存充值记录到 MySQL
	lnmcDepositHistory := &models.LnmcDepositHistory{
		Username:          username,
		WalletAddress:     walletAddress,
		BalanceLNMCBefore: int64(balanceLNMC),
		RechargeAmount:    totalAmount, //充值金额，单位是人民币
		PaymentType:       1,           //第三方支付方式 1- 支付宝， 2-微信 3-银行卡

		BalanceLNMCAfter: int64(balanceLNMC),
		BlockNumber:      blockNumber,
		TxHash:           hash,
	}

	m.AddDepositHistory(lnmcDepositHistory)

	//更新redis里用户钱包的代币余额
	redisConn.Do("HSET",
		fmt.Sprintf("userWallet:%s", username),
		"LNMCAmount",
		balanceLNMC)

	return nil
}

//数据库操作，将订单到账及退款记录到 MySQL
func (m *MysqlWalletRepository) AddLnmcOrderTransferHistory(lnmcOrderTransferHistory *models.LnmcOrderTransferHistory) error {

	if lnmcOrderTransferHistory == nil {
		return errors.New("lnmcOrderTransferHistory is nil")
	}
	//增加记录
	if err := m.db.Clauses(clause.OnConflict{DoNothing: true}).Create(lnmcOrderTransferHistory).Error; err != nil {
		m.logger.Error("增加LnmcOrderTransferHistory表失败", zap.Error(err))
		return err
	} else {
		m.logger.Debug("增加LnmcOrderTransferHistory表成功")
	}

	return nil

}

//用户注册钱包
func (m *MysqlWalletRepository) AddUserWallet(username, walletAddress, amountETHString string) (err error) {
	userWallet := &models.UserWallet{
		Username:        username,
		WalletAddress:   walletAddress,
		AmountETHString: amountETHString,
	}

	//增加记录
	if err := m.db.Clauses(clause.OnConflict{DoNothing: true}).Create(userWallet).Error; err != nil {
		m.logger.Error("增加UserWallet表失败", zap.Error(err))
		return err
	} else {
		m.logger.Debug("增加UserWallet表成功")
	}

	return nil
}

//用户充值
func (m *MysqlWalletRepository) AddDepositHistory(lnmcDepositHistory *models.LnmcDepositHistory) (err error) {

	//增加记录
	if err := m.db.Clauses(clause.OnConflict{DoNothing: true}).Create(lnmcDepositHistory).Error; err != nil {
		m.logger.Error("增加充值历史记录LnmcDepositHistory表失败", zap.Error(err))
		return err
	} else {
		m.logger.Debug("增加充值历史记录LnmcDepositHistory表成功")
	}

	return nil
}

//用户转账预审核,  新增记录
func (m *MysqlWalletRepository) AddLnmcTransferHistory(lmnccTransferHistory *models.LnmcTransferHistory) (err error) {

	//增加记录
	if err := m.db.Clauses(clause.OnConflict{DoNothing: true}).Create(lmnccTransferHistory).Error; err != nil {
		m.logger.Error("增加用户转账预审核LnmcTransferHistory表失败", zap.Error(err))
		return err
	} else {
		m.logger.Debug("增加用户转账预审核LnmcTransferHistory表成功")

	}

	return nil
}

//9-11，为某个订单支付，查询出对应的记录，然后更新 orderID, 将State修改为1
//确认转账后，更新转账历史记录
func (m *MysqlWalletRepository) UpdateLnmcTransferHistory(lmncTransferHistory *models.LnmcTransferHistory) (err error) {
	where := models.LnmcTransferHistory{
		UUID: lmncTransferHistory.UUID,
	}

	result := m.db.Model(&models.LnmcTransferHistory{}).Where(&where).Updates(lmncTransferHistory)

	//updated records count
	m.logger.Debug("UpdateLnmcTransferHistory result: ",
		zap.Int64("RowsAffected", result.RowsAffected),
		zap.Error(result.Error))

	if result.Error != nil {
		m.logger.Error("确认转账后，更新转账历史记录失败", zap.Error(result.Error))
		return result.Error
	} else {
		m.logger.Debug("确认转账后，更新转账历史记录成功")
	}

	return nil
}

//用户提现预审核,  新增记录
func (m *MysqlWalletRepository) AddLnmcWithdrawHistory(lnmcWithdrawHistory *models.LnmcWithdrawHistory) (err error) {

	//增加记录
	if err := m.db.Clauses(clause.OnConflict{DoNothing: true}).Create(lnmcWithdrawHistory).Error; err != nil {
		m.logger.Error("增加LnmcWithdrawHistory表失败", zap.Error(err))
		return err
	} else {
		m.logger.Debug("增加LnmcWithdrawHistory表成功")
	}

	return nil
}

//确认提现后，更新提现历史记录
func (m *MysqlWalletRepository) UpdateLnmcWithdrawHistory(lnmcWithdrawHistory *models.LnmcWithdrawHistory) (err error) {
	p := new(models.LnmcWithdrawHistory)
	where := models.LnmcWithdrawHistory{
		WithdrawUUID: lnmcWithdrawHistory.WithdrawUUID,
	}
	if err := m.db.Model(p).Where(&where).First(p).Error; err != nil {
		return errors.Wrapf(err, "Get lnmcWithdrawHistory error[WithdrawUUID=%s]", lnmcWithdrawHistory.WithdrawUUID)
	}
	p.State = lnmcWithdrawHistory.State
	p.BlockNumber = lnmcWithdrawHistory.BlockNumber
	p.TxHash = lnmcWithdrawHistory.TxHash
	p.BalanceLNMCBefore = lnmcWithdrawHistory.BalanceLNMCBefore
	p.AmountLNMC = lnmcWithdrawHistory.AmountLNMC
	p.BalanceLNMCAfter = lnmcWithdrawHistory.BalanceLNMCAfter

	result := m.db.Model(&models.LnmcWithdrawHistory{}).Where(&where).Updates(p)

	//updated records count
	m.logger.Debug("UpdateLnmcWithdrawHistory result: ",
		zap.Int64("RowsAffected", result.RowsAffected),
		zap.Error(result.Error))

	if result.Error != nil {
		m.logger.Error("确认提现后，更新提现历史记录失败", zap.Error(result.Error))
		return result.Error
	} else {
		m.logger.Debug("确认提现后，更新提现历史记录成功")
	}

	return nil
}

//增加接收者的收款历史表
func (m *MysqlWalletRepository) AddeCollectionHistory(lnmcCollectionHistory *models.LnmcCollectionHistory) (err error) {

	//增加记录
	if err := m.db.Clauses(clause.OnConflict{DoNothing: true}).Create(lnmcCollectionHistory).Error; err != nil {
		m.logger.Error("增加收款历史表失败", zap.Error(err))
		return err
	} else {
		m.logger.Debug("增加收款历史表成功")
	}

	return nil
}

// GetPages 分页返回数据
func (m *MysqlWalletRepository) GetPages(model interface{}, out interface{}, pageIndex, pageSize int, totalCount *int64, where interface{}, orders ...string) error {
	db2 := m.db.Model(model).Where(model).Where(where)
	if len(orders) > 0 {
		for _, order := range orders {
			db2 = db2.Order(order)
		}
	}
	err := db2.Count(totalCount).Error
	if err != nil {
		m.logger.Error("查询总数出错", zap.Error(err))
		return err
	}
	if *totalCount == 0 {
		return nil
	}
	return db2.Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(out).Error
}


//分页获取充值历史
func (m *MysqlWalletRepository) GetDepositHistorys(username string, PageNum int, PageSize int, total *int64, where interface{}) []*models.LnmcDepositHistory {
	var deposits []*models.LnmcDepositHistory
	if err := m.GetPages(&models.LnmcDepositHistory{Username: username}, &deposits, PageNum, PageSize, total, where); err != nil {
		m.logger.Error("获取充值历史失败", zap.Error(err))
	}
	return deposits
}

//分页获取提现历史
func (m *MysqlWalletRepository) GetWithdrawHistorys(username string, PageNum int, PageSize int, total *int64, where interface{}) []*models.LnmcWithdrawHistory {
	var withdraws []*models.LnmcWithdrawHistory
	if err := m.GetPages(&models.LnmcWithdrawHistory{Username: username}, &withdraws, PageNum, PageSize, total, where); err != nil {
		m.logger.Error("获取提现历史失败", zap.Error(err))
	}
	return withdraws
}

//分页获取转账历史
func (m *MysqlWalletRepository) GetTransferHistorys(username string, PageNum int, PageSize int, total *int64, where interface{}) []*models.LnmcTransferHistory {
	var transfers []*models.LnmcTransferHistory
	if err := m.GetPages(&models.LnmcTransferHistory{Username: username}, &transfers, PageNum, PageSize, total, where); err != nil {
		m.logger.Error("获取转账历史失败", zap.Error(err))
	}
	return transfers
}

//根据TxHash查询出充值记录详情
func (m *MysqlWalletRepository) GetDepositInfo(txHash string) (*models.LnmcDepositHistory, error) {

	dep := new(models.LnmcDepositHistory)

	if err := m.db.Model(dep).Where(&models.LnmcDepositHistory{
		TxHash: txHash,
	}).First(dep).Error; err != nil {
		return nil, errors.Wrapf(err, "Get LnmcDepositHistory info error[txHash=%s]", txHash)
	}
	return dep, nil
}

//根据TxHash查询出提现记录详情
func (m *MysqlWalletRepository) GetWithdrawInfo(txHash string) (*models.LnmcWithdrawHistory, error) {

	wd := new(models.LnmcWithdrawHistory)
	if err := m.db.Model(wd).Where(&models.LnmcWithdrawHistory{
		TxHash: txHash,
	}).First(wd).Error; err != nil {
		return nil, errors.Wrapf(err, "Get LnmcWithdrawHistory info error[txHash=%s]", txHash)
	}
	return wd, nil
}

//根据TxHash查询出转账记录详情
func (m *MysqlWalletRepository) GetTransferInfo(txHash string) (*models.LnmcTransferHistory, error) {

	tr := new(models.LnmcTransferHistory)
	if err := m.db.Model(tr).Where(&models.LnmcTransferHistory{
		TxHash: txHash,
	}).First(tr).Error; err != nil {
		return nil, errors.Wrapf(err, "Get LnmcTransferHistory info error[txHash=%s]", txHash)
	}
	return tr, nil
}

//根据PayType获取到VIP价格
func (m *MysqlWalletRepository) GetVipUserPrice(payType int) (*models.VipPrice, error) {
	p := new(models.VipPrice)
	where := models.VipPrice{
		PayType: payType,
	}
	if err := m.db.Model(p).Where(&where).First(p).Error; err != nil {
		return nil, errors.Wrapf(err, "PayType not found[payType=%d]", payType)
	}
	return p, nil
}

//根据productID获取到VIP价格
func (m *MysqlWalletRepository) GetVipUserPriceByProductID(productID string) (*models.VipPrice, error) {
	p := new(models.VipPrice)
	where := models.VipPrice{
		ProductID: productID,
	}
	if err := m.db.Model(p).Where(&where).First(p).Error; err != nil {
		return nil, errors.Wrapf(err, "PayType not found[productID=%s]", productID)
	}
	return p, nil
}

//会员付费成功后，按系统设定的比例进行佣金计算及写库， 需要新增3条佣金amount记录
func (s *MysqlWalletRepository) AddCommission(orderTotalAmount float64, username, orderID string) error {
	var err error
	currYearMonth := dateutil.GetYearMonthString()

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//从Distribution层级表查出所有需要分配佣金的用户账号
	distribution := new(models.Distribution)
	if err = s.db.Model(distribution).Where(&models.Distribution{
		Username: username,
	}).First(distribution).Error; err != nil {
		//记录找不到也会触发错误
		return errors.Wrapf(err, "AddCommission error or username not found")
	}

	//当商户不为空时候，则需要增加BusinessUnderling记录
	if distribution.BusinessUsername != "" {
		e := &models.BusinessUnderling{}
		if err = s.db.Model(e).Where(&models.BusinessUnderling{
			MembershipUsername: username,
			BusinessUsername:   distribution.BusinessUsername,
		}).First(e).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				s.logger.Debug("BusinessUnderling记录不存在才能添加")

				bc := &models.BusinessUnderling{
					MembershipUsername: username,                      //One Two Three
					BusinessUsername:   distribution.BusinessUsername, //归属的商户注册账号id
				}

				//增加记录
				if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&bc).Error; err != nil {
					s.logger.Error("增加BusinessUnderling失败, failed to upsert BusinessUnderling", zap.Error(err))
					return err
				} else {
					s.logger.Debug("增加BusinessUnderling成功, upsert BusinessUnderling succeed")
				}

				//增加到店铺下属用户列表 redis SADD, SMEMBERS可以获取该商户的全部下属用户总数
				storeUsersKey := fmt.Sprintf("StoreUsers:%s", distribution.BusinessUsername)
				if _, err = redisConn.Do("SADD", storeUsersKey, username); err != nil {
					s.logger.Error("SADD storelikeKey Error", zap.Error(err))
					return err
				}
			}
		}

		ee := &models.BusinessUserStatistics{}
		where := models.BusinessUserStatistics{
			BusinessUsername: distribution.BusinessUsername,
			YearMonth:        currYearMonth,
		}
		//查询出该商户的全部下属用户总数
		db2 := s.db.Model(&models.BusinessUnderling{}).Where(&where)
		var totalCount *int64
		err := db2.Count(totalCount).Error
		if err != nil {
			s.logger.Error("查询BusinessUnderling总数出错",
				zap.String("BusinessUsername", distribution.BusinessUsername),
				zap.String("YearMonth", currYearMonth),
				zap.Error(err))
			return err
		}

		if err = s.db.Model(ee).Where(&where).First(ee).Error; err != nil {
			//记录不存在, 需要添加
			if errors.Is(err, gorm.ErrRecordNotFound) {

				bcs := &models.BusinessUserStatistics{
					BusinessUsername: distribution.BusinessUsername,
					YearMonth:        currYearMonth,
					UnderlingTotal:   int64(*totalCount), //本月新增会员总数
				}

				tx2 := s.base.GetTransaction()

				if err := tx2.Create(bcs).Error; err != nil {
					s.logger.Error("增加BusinessUserStatistics失败", zap.Error(err))
					tx2.Rollback()
					return err
				}

				//提交
				tx2.Commit()
			} else {

				return errors.Wrapf(err, "Db errp")

			}
		} else {
			//记录存在, Update

			result := s.db.Model(ee).Where(&where).Update("underling_total", int64(*totalCount))
			s.logger.Debug("Update BusinessUserStatistics result: ", zap.Int64("RowsAffected", result.RowsAffected), zap.Error(result.Error))

			if result.Error != nil {
				s.logger.Error("Update BusinessUserStatistics失败", zap.Error(result.Error))
				return result.Error
			} else {
				mtxt := fmt.Sprintf("Update BusinessUserStatistics成功:  本月新增会员总数: %distribution", int64(*totalCount))
				s.logger.Debug(mtxt)
			}
		}

	}

	//支付成功后，需要插入佣金表Commission -  上级 向上第一级
	if distribution.UsernameLevelOne != "" {
		e := &models.Commission{}
		if err = s.db.Model(e).Where(&models.Commission{
			UsernameLevel: distribution.UsernameLevelOne,
			OrderID:       orderID,
		}).First(e).Error; err == nil {
			s.logger.Error("已经存在此用户佣金记录，不能新增", zap.String("UsernameLevel", distribution.UsernameLevelOne), zap.String("BusinessUsername", distribution.BusinessUsername))
			//记录不存在才能添加
			return errors.Wrapf(err, "Can not Insert Commission, because record is exists")
		}

		commissionOne := &models.Commission{
			YearMonth:        currYearMonth,
			UsernameLevel:    distribution.UsernameLevelOne,             //One Two Three
			BusinessUsername: distribution.BusinessUsername,             //归属的商户注册账号id
			Amount:           orderTotalAmount,                          //会员费用金额，单位是人民币
			OrderID:          orderID,                                   //订单ID
			Commission:       LMCommon.CommissionOne * orderTotalAmount, //TODO 第一级佣金， 按比例
		}

		//增加记录
		if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&commissionOne).Error; err != nil {
			s.logger.Error("增加commissionOne失败, failed to upsert Commission", zap.Error(err))
			return err
		} else {
			s.logger.Debug("增加commissionOne成功, upsert Commission succeed")
		}

		//用户的佣金月统计  CommissionStatistics
		nucsWhere := &models.CommissionStatistics{
			Username:  distribution.UsernameLevelOne,
			YearMonth: currYearMonth,
			IsRebate:  true, //判断是否返现
		}
		ncs := &models.CommissionStatistics{}
		if err = s.db.Model(ncs).Where(nucsWhere).First(ncs).Error; err == nil {
			s.logger.Error("CommissionStatistics表已经返现，不能新增记录 ", zap.String("YearMonth", currYearMonth), zap.String("Username", distribution.UsernameLevelOne))
		} else {

			//统计d.UsernameLevelOne对应的用户在当月的所有佣金总额
			where := models.Commission{
				UsernameLevel: distribution.UsernameLevelOne,
				YearMonth:     currYearMonth,
			}
			db2 := s.db.Model(&models.Commission{}).Where(&where)

			amount := Amount{}
			db2.Select("SUM(commission) AS total").Scan(&amount)
			s.logger.Debug("SUM统计出当月的总佣金金额",
				zap.String("username", distribution.UsernameLevelOne),
				zap.String("currYearMonth", currYearMonth),
				zap.Float64("total", amount.Total),
			)

			newnucs := models.CommissionStatistics{
				Username:        distribution.UsernameLevelOne,
				YearMonth:       currYearMonth,
				TotalCommission: amount.Total, //本月返佣总金额
				IsRebate:        false,        //默认返现的值是false
			}

			//Save
			s.db.Save(&newnucs)

		}

	}

	//支付成功后，需要插入佣金表Commission - 上上级 第二级
	if distribution.UsernameLevelTwo != "" {
		e := &models.Commission{}
		if err = s.db.Model(e).Where(&models.Commission{
			UsernameLevel: distribution.UsernameLevelTwo,
			OrderID:       orderID,
		}).First(e).Error; err == nil {
			s.logger.Error("CommissionStatistics表已经返现，不能新增记录 ", zap.String("UsernameLevel", distribution.UsernameLevelTwo), zap.String("BusinessUsername", distribution.BusinessUsername))
			//记录不存在才能添加
			return errors.Wrapf(err, "Can not Insert Commission, because record is exists")
		}

		commissionTwo := &models.Commission{
			YearMonth:        currYearMonth,
			UsernameLevel:    distribution.UsernameLevelTwo,             //One Two Three
			BusinessUsername: distribution.BusinessUsername,             //归属的商户注册账号id
			Amount:           orderTotalAmount,                          //会员费用金额，单位是人民币
			OrderID:          orderID,                                   //订单ID
			Commission:       LMCommon.CommissionTwo * orderTotalAmount, //TODO 第二级佣金
		}

		//增加记录
		if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&commissionTwo).Error; err != nil {
			s.logger.Error("增加commissionTwo失败, failed to upsert Commission", zap.Error(err))
			return err
		} else {
			s.logger.Debug("增加commissionTwo成功, upsert Commission succeed")
		}

		//用户的佣金月统计  CommissionStatistics
		nucsWhere := &models.CommissionStatistics{
			Username:  distribution.UsernameLevelTwo,
			YearMonth: currYearMonth,
			IsRebate:  true, //是否已经返现
		}
		ncs := &models.CommissionStatistics{}
		if err = s.db.Model(ncs).Where(nucsWhere).First(ncs).Error; err == nil {
			s.logger.Error("CommissionStatistics表已经返现，不能新增记录 ", zap.String("YearMonth", currYearMonth), zap.String("Username", distribution.UsernameLevelTwo))
		} else {
			//统计d.UsernameLevelTwo对应的用户在当月的所有佣金总额
			where := models.Commission{
				UsernameLevel: distribution.UsernameLevelTwo,
				YearMonth:     currYearMonth,
			}
			db2 := s.db.Model(&models.Commission{}).Where(&where)
			amount := Amount{}
			db2.Select("SUM(commission) AS total").Scan(&amount)

			newnucs := &models.CommissionStatistics{
				Username:        distribution.UsernameLevelTwo,
				YearMonth:       currYearMonth,
				TotalCommission: amount.Total, //本月返佣总金额
				IsRebate:        false,        //默认返现的值是false
			}
			//Save
			s.db.Save(&newnucs)
		}

	}

	//支付成功后，需要插入佣金表Commission - 上上上级 向上第三级
	if distribution.UsernameLevelThree != "" {
		e := &models.Commission{}
		if err = s.db.Model(e).Where(&models.Commission{
			UsernameLevel: distribution.UsernameLevelThree,
			OrderID:       orderID,
		}).First(e).Error; err == nil {
			s.logger.Error("CommissionStatistics表已经返现，不能新增记录 ", zap.String("UsernameLevel", distribution.UsernameLevelThree), zap.String("BusinessUsername", distribution.BusinessUsername))
			//记录不存在才能添加
			return errors.Wrapf(err, "Can not Insert Commission, because record is exists")
		}
		commissionThree := &models.Commission{
			YearMonth:        currYearMonth,
			UsernameLevel:    distribution.UsernameLevelThree,             //One Two Three
			BusinessUsername: distribution.BusinessUsername,               //归属的商户注册账号id
			Amount:           orderTotalAmount,                            //会员费用金额，单位是人民币
			OrderID:          orderID,                                     //订单ID
			Commission:       LMCommon.CommissionThree * orderTotalAmount, //TODO 第三级佣金
		}

		//增加记录
		if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&commissionThree).Error; err != nil {
			s.logger.Error("增加commissionThree失败, failed to upsert Commission", zap.Error(err))
			return err
		} else {
			s.logger.Debug("增加commissionThree成功, upsert Commission succeed")
		}

		//用户的佣金月统计  CommissionStatistics
		nucs := &models.CommissionStatistics{
			Username:  distribution.UsernameLevelThree,
			YearMonth: currYearMonth,
		}
		ncs := &models.CommissionStatistics{}
		if err = s.db.Model(ncs).Where(nucs).First(ncs).Error; err == nil {
			s.logger.Error("CommissionStatistics表已经返现，不能新增记录 ", zap.String("YearMonth", currYearMonth), zap.String("UsernameLevelThree", distribution.UsernameLevelThree))
		} else {
			//统计d.UsernameLevelThree对应的用户在当月的所有佣金总额
			where := models.Commission{
				UsernameLevel: distribution.UsernameLevelThree,
				YearMonth:     currYearMonth,
			}
			db2 := s.db.Model(&models.Commission{}).Where(&where)
			amount := Amount{}
			db2.Select("SUM(commission) AS total").Scan(&amount)

			newnucs := &models.CommissionStatistics{
				Username:        distribution.UsernameLevelThree,
				YearMonth:       currYearMonth,
				TotalCommission: amount.Total, //本月返佣总金额
				IsRebate:        false,        //默认返现的值是false
			}
			//Save
			s.db.Save(&newnucs)
		}

	}

	return nil
}

//购买Vip后，增加用户的到期时间
func (s *MysqlWalletRepository) AddVipEndDate(username string, endTime int64) error {
	user := new(models.User)

	user.UserBase.VipEndDate = endTime
	user.UserBase.State = 1 //1-Vip用户

	where := models.User{
		UserBase: models.UserBase{
			Username: username,
		},
	}

	// 同时更新多个字段
	result := s.db.Model(&models.User{}).Where(&where).Select("vip_end_date", "state").Updates(user)

	//updated records counts
	s.logger.Debug("AddVipEndDate result: ",
		zap.Int64("RowsAffected", result.RowsAffected),
		zap.Error(result.Error))

	if result.Error != nil {
		s.logger.Error("增加用户的到期时间及Vip状态失败", zap.Error(result.Error))
		return result.Error
	} else {
		s.logger.Debug("增加用户的到期时间及Vip状态成功")
	}

	return nil

}

// 修改ChargeHistory的支付及哈希区块
func (s *MysqlWalletRepository) UpdateChargeHistoryForPayed(chargeHistory *models.ChargeHistory) error {
	values := map[string]interface{}{"is_payed": true, "block_number": chargeHistory.BlockNumber, "tx_hash": chargeHistory.TxHash}
	result := s.db.Model(&models.ChargeHistory{}).Where(&models.ChargeHistory{
		ChargeOrderID: chargeHistory.ChargeOrderID,
	}).Updates(values)

	//updated records counts
	s.logger.Debug("UpdateChargeHistoryForPayed result: ",
		zap.Int64("RowsAffected", result.RowsAffected),
		zap.Error(result.Error))

	if result.Error != nil {
		s.logger.Error("修改ChargeHistory的支付及哈希区块失败", zap.Error(result.Error))
		return result.Error
	} else {
		s.logger.Debug("修改ChargeHistory的支付及哈希区块成功")
	}

	return nil
}

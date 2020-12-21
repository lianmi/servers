package repositories

import (
	"context"
	"fmt"
	"github.com/gomodule/redigo/redis"
	Wallet "github.com/lianmi/servers/api/proto/wallet"
	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"github.com/smartwalle/alipay/v3"
	"github.com/smartwalle/xid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	// "strconv"
	uuid "github.com/satori/go.uuid"
)

type WalletRepository interface {
	DoPreAlipay(ctx context.Context, req *Wallet.PreAlipayReq) (*Wallet.PreAlipayResp, error)

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

	GetCollectionHistorys(toUsername, fromUsername string, PageNum int, PageSize int, total *int64, where interface{}) []*models.LnmcCollectionHistory

	GetDepositHistorys(username string, PageNum int, PageSize int, total *int64, where interface{}) []*models.LnmcDepositHistory

	GetWithdrawHistorys(username string, PageNum int, PageSize int, total *int64, where interface{}) []*models.LnmcWithdrawHistory

	GetTransferHistorys(username string, PageNum int, PageSize int, total *int64, where interface{}) []*models.LnmcTransferHistory

	GetDepositInfo(txHash string) (*models.LnmcDepositHistory, error)

	GetWithdrawInfo(txHash string) (*models.LnmcWithdrawHistory, error)

	GetTransferInfo(txHash string) (*models.LnmcTransferHistory, error)

	//根据PayType获取到VIP价格
	GetVipUserPrice(payType int) (*models.VipPrice, error)
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
	} else {

	}

	walletAddress, err = redis.String(redisConn.Do("HGET", fmt.Sprintf("userWallet:%s", username), "WalletAddress"))
	if err != nil {
		m.logger.Error("HGET失败", zap.Error(result.Error))
		return err
	}

	//保存充值记录到 MySQL
	lnmcDepositHistory := &models.LnmcDepositHistory{
		UUID:              uuid.NewV4().String(),
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
	//如果没有记录，则增加，如果有记录，则更新全部字段
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

	//如果没有记录，则增加，如果有记录，则更新全部字段
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

	//如果没有记录，则增加，如果有记录，则更新全部字段
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

	//如果没有记录，则增加，如果有记录，则更新全部字段
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

	//如果没有记录，则增加，如果有记录，则更新全部字段
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

	//如果没有记录，则增加，如果有记录，则更新全部字段
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
	db := m.db.Model(model).Where(model)
	db = db.Where(where)
	if len(orders) > 0 {
		for _, order := range orders {
			db = db.Order(order)
		}
	}
	err := db.Count(totalCount).Error
	if err != nil {
		m.logger.Error("查询总数出错", zap.Error(err))
		return err
	}
	if *totalCount == 0 {
		return nil
	}
	return db.Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(out).Error
}

//分页获取收款历史
func (m *MysqlWalletRepository) GetCollectionHistorys(toUsername, fromUsername string, PageNum int, PageSize int, total *int64, where interface{}) []*models.LnmcCollectionHistory {
	var collections []*models.LnmcCollectionHistory
	if fromUsername == "" {
		if err := m.GetPages(&models.LnmcCollectionHistory{ToUsername: toUsername}, &collections, PageNum, PageSize, total, where); err != nil {
			m.logger.Error("获取收款历史失败", zap.Error(err))
		}
	} else {
		if err := m.GetPages(&models.LnmcCollectionHistory{ToUsername: toUsername, FromUsername: fromUsername}, &collections, PageNum, PageSize, total, where); err != nil {
			m.logger.Error("获取收款历史失败", zap.Error(err))
		}
	}

	return collections
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

package repositories

import (
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type WalletRepository interface {
	SaveLnmcOrderTransferHistory(lnmcOrderTransferHistory *models.LnmcOrderTransferHistory) error

	SaveUserWallet(username, walletAddress, amountETHString string) (err error)

	SaveDepositHistory(lnmcDepositHistory *models.LnmcDepositHistory) (err error)

	SaveLnmcTransferHistory(lmnccTransferHistory *models.LnmcTransferHistory) (err error)

	UpdateLnmcTransferHistory(lmncTransferHistory *models.LnmcTransferHistory) (err error)

	SaveLnmcWithdrawHistory(lnmcWithdrawHistory *models.LnmcWithdrawHistory) (err error)

	UpdateLnmcWithdrawHistory(lnmcWithdrawHistory *models.LnmcWithdrawHistory) (err error)

	SaveCollectionHistory(lnmcCollectionHistory *models.LnmcCollectionHistory) (err error)

	GetPages(model interface{}, out interface{}, pageIndex, pageSize int, totalCount *uint64, where interface{}, orders ...string) error

	GetCollectionHistorys(toUsername, fromUsername string, PageNum int, PageSize int, total *uint64, where interface{}) []*models.LnmcCollectionHistory

	GetDepositHistorys(username string, PageNum int, PageSize int, total *uint64, where interface{}) []*models.LnmcDepositHistory

	GetWithdrawHistorys(username string, PageNum int, PageSize int, total *uint64, where interface{}) []*models.LnmcWithdrawHistory

	GetTransferHistorys(username string, PageNum int, PageSize int, total *uint64, where interface{}) []*models.LnmcTransferHistory

	GetDepositInfo(txHash string) (*models.LnmcDepositHistory, error)

	GetWithdrawInfo(txHash string) (*models.LnmcWithdrawHistory, error)

	GetTransferInfo(txHash string) (*models.LnmcTransferHistory, error)
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

//数据库操作，将订单到账及退款记录到 MySQL
func (m *MysqlWalletRepository) SaveLnmcOrderTransferHistory(lnmcOrderTransferHistory *models.LnmcOrderTransferHistory) error {
	tx := m.base.GetTransaction()

	if err := tx.Save(lnmcOrderTransferHistory).Error; err != nil {
		m.logger.Error("更新订单到账及退款记录失败", zap.Error(err))
		tx.Rollback()
		return err

	}

	//提交
	tx.Commit()

	return nil

}

//用户注册钱包
func (m *MysqlWalletRepository) SaveUserWallet(username, walletAddress, amountETHString string) (err error) {
	userWallet := &models.UserWallet{
		Username:        username,
		WalletAddress:   walletAddress,
		AmountETHString: amountETHString,
	}

	tx := m.base.GetTransaction()

	if err := tx.Save(userWallet).Error; err != nil {
		m.logger.Error("更新UserWallet失败", zap.Error(err))
		tx.Rollback()
		return err

	}
	//提交
	tx.Commit()

	return nil
}

//用户充值
func (m *MysqlWalletRepository) SaveDepositHistory(lnmcDepositHistory *models.LnmcDepositHistory) (err error) {

	tx := m.base.GetTransaction()

	if err := tx.Save(lnmcDepositHistory).Error; err != nil {
		m.logger.Error("更新充值历史记录表失败", zap.Error(err))
		tx.Rollback()
		return err

	}
	//提交
	tx.Commit()

	return nil
}

//用户转账预审核,  新增记录
func (m *MysqlWalletRepository) SaveLnmcTransferHistory(lmnccTransferHistory *models.LnmcTransferHistory) (err error) {

	tx := m.base.GetTransaction()

	if err := tx.Save(lmnccTransferHistory).Error; err != nil {
		m.logger.Error("更新用户转账预审核表失败", zap.Error(err))
		tx.Rollback()
		return err

	}
	//提交
	tx.Commit()

	return nil
}

//9-11，为某个订单支付，查询出对应的记录，然后更新 orderID 及 signedTx, 将State修改为1
//确认转账后，更新转账历史记录
func (m *MysqlWalletRepository) UpdateLnmcTransferHistory(lmncTransferHistory *models.LnmcTransferHistory) (err error) {
	p := new(models.LnmcTransferHistory)
	if err := m.db.Model(p).Where("username = ? and to_username=?", lmncTransferHistory.Username, lmncTransferHistory.ToUsername).First(p).Error; err != nil {
		return errors.Wrapf(err, "Get LnmcTransferHistory error")
	}
	p.State = lmncTransferHistory.State
	// p.SignedTx = lmncTransferHistory.SignedTx
	p.BlockNumber = lmncTransferHistory.BlockNumber
	p.TxHash = lmncTransferHistory.TxHash
	if lmncTransferHistory.OrderID != "" {
		p.OrderID = lmncTransferHistory.OrderID
	}
	p.BalanceLNMCBefore = lmncTransferHistory.BalanceLNMCBefore
	p.AmountLNMC = lmncTransferHistory.AmountLNMC
	p.BalanceLNMCAfter = lmncTransferHistory.BalanceLNMCAfter

	tx := m.base.GetTransaction()

	if err := tx.Save(p).Error; err != nil {
		m.logger.Error("更新用户转账预审核表失败", zap.Error(err))
		tx.Rollback()
		return err
	}
	//提交
	tx.Commit()

	return nil
}

//用户提现预审核,  新增记录
func (m *MysqlWalletRepository) SaveLnmcWithdrawHistory(lnmcWithdrawHistory *models.LnmcWithdrawHistory) (err error) {

	tx := m.base.GetTransaction()

	if err := tx.Save(lnmcWithdrawHistory).Error; err != nil {
		m.logger.Error("更新用户提现预审核表失败", zap.Error(err))
		tx.Rollback()
		return err

	}
	//提交
	tx.Commit()

	return nil
}

//确认提现后，更新提现历史记录
func (m *MysqlWalletRepository) UpdateLnmcWithdrawHistory(lnmcWithdrawHistory *models.LnmcWithdrawHistory) (err error) {
	p := new(models.LnmcWithdrawHistory)
	if err := m.db.Model(p).Where("withdraw_uuid = ?", lnmcWithdrawHistory.WithdrawUUID).First(p).Error; err != nil {
		return errors.Wrapf(err, "Get lnmcWithdrawHistory error[WithdrawUUID=%s]", lnmcWithdrawHistory.WithdrawUUID)
	}
	p.State = lnmcWithdrawHistory.State
	p.BlockNumber = lnmcWithdrawHistory.BlockNumber
	p.TxHash = lnmcWithdrawHistory.TxHash
	p.BalanceLNMCBefore = lnmcWithdrawHistory.BalanceLNMCBefore
	p.AmountLNMC = lnmcWithdrawHistory.AmountLNMC
	p.BalanceLNMCAfter = lnmcWithdrawHistory.BalanceLNMCAfter

	tx := m.base.GetTransaction()

	if err := tx.Save(p).Error; err != nil {
		m.logger.Error("更新用户提现表失败", zap.Error(err))
		tx.Rollback()
		return err
	}
	//提交
	tx.Commit()

	return nil
}

//收款历史表
func (m *MysqlWalletRepository) SaveCollectionHistory(lnmcCollectionHistory *models.LnmcCollectionHistory) (err error) {

	tx := m.base.GetTransaction()

	if err := tx.Save(lnmcCollectionHistory).Error; err != nil {
		m.logger.Error("更新收款历史记录表失败", zap.Error(err))
		tx.Rollback()
		return err

	}
	//提交
	tx.Commit()

	return nil
}

// GetPages 分页返回数据
func (m *MysqlWalletRepository) GetPages(model interface{}, out interface{}, pageIndex, pageSize int, totalCount *uint64, where interface{}, orders ...string) error {
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
func (m *MysqlWalletRepository) GetCollectionHistorys(toUsername, fromUsername string, PageNum int, PageSize int, total *uint64, where interface{}) []*models.LnmcCollectionHistory {
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
func (m *MysqlWalletRepository) GetDepositHistorys(username string, PageNum int, PageSize int, total *uint64, where interface{}) []*models.LnmcDepositHistory {
	var deposits []*models.LnmcDepositHistory
	if err := m.GetPages(&models.LnmcDepositHistory{Username: username}, &deposits, PageNum, PageSize, total, where); err != nil {
		m.logger.Error("获取充值历史失败", zap.Error(err))
	}
	return deposits
}

//分页获取提现历史
func (m *MysqlWalletRepository) GetWithdrawHistorys(username string, PageNum int, PageSize int, total *uint64, where interface{}) []*models.LnmcWithdrawHistory {
	var withdraws []*models.LnmcWithdrawHistory
	if err := m.GetPages(&models.LnmcWithdrawHistory{Username: username}, &withdraws, PageNum, PageSize, total, where); err != nil {
		m.logger.Error("获取提现历史失败", zap.Error(err))
	}
	return withdraws
}

//分页获取转账历史
func (m *MysqlWalletRepository) GetTransferHistorys(username string, PageNum int, PageSize int, total *uint64, where interface{}) []*models.LnmcTransferHistory {
	var transfers []*models.LnmcTransferHistory
	if err := m.GetPages(&models.LnmcTransferHistory{Username: username}, &transfers, PageNum, PageSize, total, where); err != nil {
		m.logger.Error("获取转账历史失败", zap.Error(err))
	}
	return transfers
}

//根据TxHash查询出充值记录详情
func (m *MysqlWalletRepository) GetDepositInfo(txHash string) (*models.LnmcDepositHistory, error) {

	dep := new(models.LnmcDepositHistory)
	if err := m.db.Model(dep).Where("tx_hash = ?", txHash).First(dep).Error; err != nil {
		return nil, errors.Wrapf(err, "Get LnmcDepositHistory info error[txHash=%s]", txHash)
	}
	return dep, nil
}

//根据TxHash查询出提现记录详情
func (m *MysqlWalletRepository) GetWithdrawInfo(txHash string) (*models.LnmcWithdrawHistory, error) {

	wd := new(models.LnmcWithdrawHistory)
	if err := m.db.Model(wd).Where("tx_hash = ?", txHash).First(wd).Error; err != nil {
		return nil, errors.Wrapf(err, "Get LnmcWithdrawHistory info error[txHash=%s]", txHash)
	}
	return wd, nil
}

//根据TxHash查询出转账记录详情
func (m *MysqlWalletRepository) GetTransferInfo(txHash string) (*models.LnmcTransferHistory, error) {

	tr := new(models.LnmcTransferHistory)
	if err := m.db.Model(tr).Where("tx_hash = ?", txHash).First(tr).Error; err != nil {
		return nil, errors.Wrapf(err, "Get LnmcTransferHistory info error[txHash=%s]", txHash)
	}
	return tr, nil
}

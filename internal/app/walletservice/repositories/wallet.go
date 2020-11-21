package repositories

import (
	"github.com/gomodule/redigo/redis"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WalletRepository interface {
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

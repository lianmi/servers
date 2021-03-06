package repositories

import (
	"github.com/lianmi/servers/internal/pkg/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

//用户类型 1-普通
func IsNormalUser(db *gorm.DB) *gorm.DB {
	return db.Where("user_type = ? ", 1)
}

//用户类型 2-商户  处于预审核状态
func IsPreBusinessUser(db *gorm.DB) *gorm.DB {
	return db.Where("user_type = ?  and  state = ?", 2, 0)
}

//用户类型 2-商户  处于已审核状态
func IsBusinessUser(db *gorm.DB) *gorm.DB {
	return db.Where("user_type = ?  and  state = ?", 2, 1)
}

//按createAt的时间段
func BetweenCreatedAt(startAt, endAt uint64) func(db *gorm.DB) *gorm.DB {
	if endAt > 0 {
		return func(db *gorm.DB) *gorm.DB {
			return db.Where("created_at>= ? and created_at<= ? ", startAt, endAt)
		}
	} else {
		return func(db *gorm.DB) *gorm.DB {
			return db
		}
	}
}

//按createAt的时间段
func BetweenUpdatedAt(startAt, endAt uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("updated_at>= ? and updated_at<= ? ", startAt, endAt)
	}
}

//实名
func TrueName(trueNames []string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("true_name IN (?)", trueNames)
	}
}

//法人
func LegalPerson(legalPersons []string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("legal_person IN (?)", legalPersons)
	}
}

//店铺名称
func Branchesname(branchesnames []string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("branchesname IN (?)", branchesnames)
	}
}

func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page == 0 {
			page = 1
		}

		switch {
		case pageSize > 100:
			pageSize = 20
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

//分页查询lnmc_collection_histories列表, pageSize default 20
func (s *MysqlWalletRepository) GetCollectionHistorys(toUsername, fromUsername string, startAt, endAt uint64, pageNum int, pageSize int, total *int64) ([]*models.LnmcCollectionHistory, error) {

	var count int64
	var lnmcCollectionHistories []*models.LnmcCollectionHistory
	model := new(models.LnmcCollectionHistory)

	s.db.Model(&model).Find(&lnmcCollectionHistories, "to_username = ? AND from_username=?", toUsername, fromUsername)

	resultError := s.db.Model(&model).Scopes(BetweenCreatedAt(startAt, endAt), Paginate(pageNum, pageSize)).Order("created_at DESC").Find(&lnmcCollectionHistories)

	count = int64(len(lnmcCollectionHistories))
	total = &count
	s.logger.Debug("符合条件的记录总数", zap.Int64("count", count))

	for _, history := range lnmcCollectionHistories {
		s.logger.Debug("分页显示lnmc_collection_histories列表 ",
			zap.Int64("CreatedAt", history.CreatedAt),
			zap.String("ToUsername", history.ToUsername),
			zap.String("FromUsername", history.FromUsername),
			zap.String("OrderID", history.OrderID),
		)
	}
	if resultError.Error != nil {
		return nil, resultError.Error
	}

	return lnmcCollectionHistories, nil

}

//分页获取转账历史
func (s *MysqlWalletRepository) GetTransferHistorys(toUsername string, startAt, endAt uint64, pageNum int, pageSize int, total *int64) ([]*models.LnmcTransferHistory, error) {
	var transfers []*models.LnmcTransferHistory

	var count int64
	model := new(models.LnmcTransferHistory)

	s.db.Model(&model).Find(&transfers, "to_username = ?", toUsername)

	resultError := s.db.Model(&model).Scopes(BetweenCreatedAt(startAt, endAt), Paginate(pageNum, pageSize)).Order("created_at DESC").Find(&transfers)

	count = int64(len(transfers))
	total = &count
	s.logger.Debug("符合条件的记录总数 ", zap.Int64("count", count))

	for _, history := range transfers {
		s.logger.Debug("分页显示lnmc_transfer_histories列表 ",
			zap.Int64("CreatedAt", history.CreatedAt),
			zap.String("ToUsername", history.ToUsername),
			zap.String("OrderID", history.OrderID),
		)
	}
	if resultError.Error != nil {
		return nil, resultError.Error
	}

	return transfers, nil
}

//分页获取充值历史
func (s *MysqlWalletRepository) GetDepositHistorys(username string, startAt, endAt uint64, pageNum int, pageSize int, total *int64) ([]*models.LnmcDepositHistory, error) {
	var deposits []*models.LnmcDepositHistory
	var count int64
	model := new(models.LnmcDepositHistory)

	s.db.Model(&model).Find(&deposits, "username = ?", username)

	resultError := s.db.Model(&model).Scopes(BetweenCreatedAt(startAt, endAt), Paginate(pageNum, pageSize)).Order("created_at DESC").Find(&deposits)

	count = int64(len(deposits))
	total = &count
	s.logger.Debug("符合条件的记录总数 ", zap.Int64("count", count))

	for _, history := range deposits {
		s.logger.Debug("分页显示lnmc_deposit_histories列表 ",
			zap.Int64("CreatedAt", history.CreatedAt),
			zap.String("Username", history.Username),
			zap.String("WalletAddress", history.WalletAddress),
			zap.Float64("RechargeAmount", history.RechargeAmount),
			zap.Int64("BalanceLNMCBefore", history.BalanceLNMCBefore),
			zap.Int64("BalanceLNMCAfter", history.BalanceLNMCAfter),
		)
	}
	if resultError.Error != nil {
		return nil, resultError.Error
	}

	return deposits, nil
}

//分页获取提现历史
func (s *MysqlWalletRepository) GetWithdrawHistorys(username string, startAt, endAt uint64, pageNum int, pageSize int, total *int64) ([]*models.LnmcWithdrawHistory, error) {
	var withdraws []*models.LnmcWithdrawHistory

	var count int64
	model := new(models.LnmcWithdrawHistory)

	s.db.Model(&model).Find(&withdraws, "username = ?", username)

	resultError := s.db.Model(&model).Scopes(BetweenCreatedAt(startAt, endAt), Paginate(pageNum, pageSize)).Order("created_at DESC").Find(&withdraws)

	count = int64(len(withdraws))
	total = &count
	s.logger.Debug("符合条件的记录总数 ", zap.Int64("count", count))

	for _, history := range withdraws {
		s.logger.Debug("分页显示lnmc_withdraw_histories列表 ",
			zap.Int64("CreatedAt", history.CreatedAt),
			zap.String("Username", history.Username),
			zap.String("Bank", history.Bank),
			zap.String("BankCard", history.BankCard),
			zap.String("CardOwner", history.CardOwner),
			zap.Uint64("AmountLNMC", history.AmountLNMC),
		)
	}
	if resultError.Error != nil {
		return nil, resultError.Error
	}

	return withdraws, nil
}

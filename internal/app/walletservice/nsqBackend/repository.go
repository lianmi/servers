package nsqBackend

import (
	"github.com/jinzhu/gorm"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

//GetTransaction 获取事务
func (nc *NsqClient) GetTransaction() *gorm.DB {
	return nc.db.Begin()
}

//用户注册钱包
func (nc *NsqClient) SaveUserWallet(username, walletAddress, amountETHString string) (err error) {
	userWallet := &models.UserWallet{
		Username:        username,
		WalletAddress:   walletAddress,
		AmountETHString: amountETHString,
	}

	tx := nc.GetTransaction()

	if err := tx.Save(userWallet).Error; err != nil {
		nc.logger.Error("更新UserWallet失败", zap.Error(err))
		tx.Rollback()
		return err

	}
	//提交
	tx.Commit()

	return nil
}

//用户充值
func (nc *NsqClient) SaveDepositHistory(lnmcDepositHistory *models.LnmcDepositHistory) (err error) {

	tx := nc.GetTransaction()

	if err := tx.Save(lnmcDepositHistory).Error; err != nil {
		nc.logger.Error("更新充值历史记录表失败", zap.Error(err))
		tx.Rollback()
		return err

	}
	//提交
	tx.Commit()

	return nil
}

//用户转账预审核,  新增记录
func (nc *NsqClient) SaveLnmcTransferHistory(lmnccTransferHistory *models.LnmcTransferHistory) (err error) {

	tx := nc.GetTransaction()

	if err := tx.Save(lmnccTransferHistory).Error; err != nil {
		nc.logger.Error("更新用户转账预审核表失败", zap.Error(err))
		tx.Rollback()
		return err

	}
	//提交
	tx.Commit()

	return nil
}

//9-11，为某个订单支付，查询出contractAddress对应的记录，然后更新 orderID 及 signedTx, 将State修改为1
//确认转账后，更新转账历史记录
func (nc *NsqClient) UpdateLnmcTransferHistory(lmncTransferHistory *models.LnmcTransferHistory) (err error) {
	p := new(models.LnmcTransferHistory)
	if err := nc.db.Model(p).Where("username = ? and to_username=?", lmncTransferHistory.Username, lmncTransferHistory.ToUsername).First(p).Error; err != nil {
		return errors.Wrapf(err, "Get LnmcTransferHistory error")
	}
	p.State = lmncTransferHistory.State
	// p.SignedTx = lmncTransferHistory.SignedTx
	p.BlockNumber = lmncTransferHistory.BlockNumber
	p.Hash = lmncTransferHistory.Hash
	if lmncTransferHistory.OrderID != "" {
		p.OrderID = lmncTransferHistory.OrderID
	}
	p.BalanceLNMCBefore = lmncTransferHistory.BalanceLNMCBefore
	p.AmountLNMC = lmncTransferHistory.AmountLNMC
	p.BalanceLNMCAfter = lmncTransferHistory.BalanceLNMCAfter

	tx := nc.GetTransaction()

	if err := tx.Save(p).Error; err != nil {
		nc.logger.Error("更新用户转账预审核表失败", zap.Error(err))
		tx.Rollback()
		return err
	}
	//提交
	tx.Commit()

	return nil
}

//用户提现预审核,  新增记录
func (nc *NsqClient) SaveLnmcWithdrawHistory(lnmcWithdrawHistory *models.LnmcWithdrawHistory) (err error) {

	tx := nc.GetTransaction()

	if err := tx.Save(lnmcWithdrawHistory).Error; err != nil {
		nc.logger.Error("更新用户提现预审核表失败", zap.Error(err))
		tx.Rollback()
		return err

	}
	//提交
	tx.Commit()

	return nil
}

//确认提现后，更新提现历史记录
func (nc *NsqClient) UpdateLnmcWithdrawHistory(lnmcWithdrawHistory *models.LnmcWithdrawHistory) (err error) {
	p := new(models.LnmcWithdrawHistory)
	if err := nc.db.Model(p).Where("withdraw_uuid = ?", lnmcWithdrawHistory.WithdrawUUID).First(p).Error; err != nil {
		return errors.Wrapf(err, "Get lnmcWithdrawHistory error[WithdrawUUID=%s]", lnmcWithdrawHistory.WithdrawUUID)
	}
	p.State = lnmcWithdrawHistory.State
	p.BlockNumber = lnmcWithdrawHistory.BlockNumber
	p.Hash = lnmcWithdrawHistory.Hash
	p.BalanceLNMCBefore = lnmcWithdrawHistory.BalanceLNMCBefore
	p.AmountLNMC = lnmcWithdrawHistory.AmountLNMC
	p.BalanceLNMCAfter = lnmcWithdrawHistory.BalanceLNMCAfter

	tx := nc.GetTransaction()

	if err := tx.Save(p).Error; err != nil {
		nc.logger.Error("更新用户提现表失败", zap.Error(err))
		tx.Rollback()
		return err
	}
	//提交
	tx.Commit()

	return nil
}

//收款历史表
func (nc *NsqClient) SaveCollectionHistory(lnmcCollectionHistory *models.LnmcCollectionHistory) (err error) {

	tx := nc.GetTransaction()

	if err := tx.Save(lnmcCollectionHistory).Error; err != nil {
		nc.logger.Error("更新收款历史记录表失败", zap.Error(err))
		tx.Rollback()
		return err

	}
	//提交
	tx.Commit()

	return nil
}

// GetPages 分页返回数据
func (nc *NsqClient) GetPages(model interface{}, out interface{}, pageIndex, pageSize int, totalCount *uint64, where interface{}, orders ...string) error {
	db := nc.db.Model(model).Where(model)
	db = db.Where(where)
	if len(orders) > 0 {
		for _, order := range orders {
			db = db.Order(order)
		}
	}
	err := db.Count(totalCount).Error
	if err != nil {
		nc.logger.Error("查询总数出错", zap.Error(err))
		return err
	}
	if *totalCount == 0 {
		return nil
	}
	return db.Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(out).Error
}

//分页获取收款历史
func (nc *NsqClient) GetCollectionHistorys(PageNum int, PageSize int, total *uint64, where interface{}) []*models.LnmcCollectionHistory {
	var collections []*models.LnmcCollectionHistory
	if err := nc.GetPages(&models.LnmcCollectionHistory{}, &collections, PageNum, PageSize, total, where); err != nil {
		nc.logger.Error("获取收款历史失败", zap.Error(err))
	}
	return collections
}

//分页获取充值历史
func (nc *NsqClient) GetDepositHistorys(PageNum int, PageSize int, total *uint64, where interface{}) []*models.LnmcDepositHistory {
	var deposits []*models.LnmcDepositHistory
	if err := nc.GetPages(&models.LnmcDepositHistory{}, &deposits, PageNum, PageSize, total, where); err != nil {
		nc.logger.Error("获取充值历史失败", zap.Error(err))
	}
	return deposits
}

//分页获取提现历史
func (nc *NsqClient) GetWithdrawHistorys(PageNum int, PageSize int, total *uint64, where interface{}) []*models.LnmcWithdrawHistory {
	var withdraws []*models.LnmcWithdrawHistory
	if err := nc.GetPages(&models.LnmcWithdrawHistory{}, &withdraws, PageNum, PageSize, total, where); err != nil {
		nc.logger.Error("获取提现历史失败", zap.Error(err))
	}
	return withdraws
}

//分页获取转账历史
func (nc *NsqClient) GetTransferHistorys(PageNum int, PageSize int, total *uint64, where interface{}) []*models.LnmcTransferHistory {
	var transfers []*models.LnmcTransferHistory
	if err := nc.GetPages(&models.LnmcTransferHistory{}, &transfers, PageNum, PageSize, total, where); err != nil {
		nc.logger.Error("获取转账历史失败", zap.Error(err))
	}
	return transfers
}

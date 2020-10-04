package nsqBackend

import (
	"github.com/jinzhu/gorm"
	"github.com/lianmi/servers/internal/pkg/models"
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
func (nc *NsqClient) UpdateLnmcTransferHistoryByContractAddress(contractAddress, orderID, SignedTx string) (err error) {

	// tx := nc.GetTransaction()

	// if err := tx.Save(lmnccTransferHistory).Error; err != nil {
	// 	nc.logger.Error("更新用户转账预审核表失败", zap.Error(err))
	// 	tx.Rollback()
	// 	return err

	// }
	// //提交
	// tx.Commit()

	return nil
}

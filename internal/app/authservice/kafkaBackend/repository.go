package kafkaBackend

import (
	"github.com/lianmi/servers/internal/pkg/models"
	"go.uber.org/zap"
)

//更新好友
func (kc *KafkaClient) SaveAddFriend(pFriend *models.Friend) error {
	//使用事务同时更新好友数据
	tx := kc.GetTransaction()

	if err := tx.Save(pFriend).Error; err != nil {
		kc.logger.Error("更新好友表失败", zap.Error(err))
		tx.Rollback()

	}
	//提交
	tx.Commit()

	return nil
}

//删除好友
func (kc *KafkaClient) DeleteFriend(userID, friendUserID uint64) error {
	where := models.Friend{UserID: userID, FriendUserID: friendUserID}
	db := kc.db.Where(&where).Delete(models.Friend{})
	err := db.Error
	if err != nil {
		kc.logger.Error("DeleteFriend出错", zap.Error(err))
		return err
	}
	count := db.RowsAffected
	kc.logger.Debug("DeleteFriend成功", zap.Int64("count", count))
	return nil
}

//创建群
func (kc *KafkaClient) SaveCreateTeam(pTeam *models.Team) error {
	//使用事务同时更新创建群数据
	tx := kc.GetTransaction()

	if err := tx.Save(pTeam).Error; err != nil {
		kc.logger.Error("更新群team表失败", zap.Error(err))
		tx.Rollback()
	}
	
	//提交
	tx.Commit()

	return nil
}

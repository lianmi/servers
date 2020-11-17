package repositories

import (
	"github.com/lianmi/servers/internal/pkg/models"
	"go.uber.org/zap"
)

//更新好友
func (s *MysqlLianmiRepository) SaveFriend(pFriend *models.Friend) error {
	//使用事务同时更新好友数据
	tx := s.base.GetTransaction()

	if err := tx.Save(pFriend).Error; err != nil {
		s.logger.Error("更新好友表失败", zap.Error(err))
		tx.Rollback()

	}
	//提交
	tx.Commit()

	return nil
}

//删除好友
func (s *MysqlLianmiRepository) DeleteFriend(userID, friendUserID uint64) error {
	where := models.Friend{UserID: userID, FriendUserID: friendUserID}
	db := s.db.Where(&where).Delete(models.Friend{})
	err := db.Error
	if err != nil {
		s.logger.Error("DeleteFriend出错", zap.Error(err))
		return err
	}
	count := db.RowsAffected
	s.logger.Debug("DeleteFriend成功", zap.Int64("count", count))
	return nil
}

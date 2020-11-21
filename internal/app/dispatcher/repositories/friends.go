package repositories

import (
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"
)

// 增加好友
func (s *MysqlLianmiRepository) AddFriend(pFriend *models.Friend) error {

	if pFriend == nil {
		return errors.New("pFriend is nil")
	}
	//如果没有记录，则增加，如果有记录，则更新全部字段
	if err := s.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&pFriend).Error; err != nil {
		s.logger.Error("AddFriend, failed to upsert friends", zap.Error(err))
		return err
	} else {
		s.logger.Debug("AddFriend, upsert friends succeed")
	}

	return nil
}

//修改好友资料
func (s *MysqlLianmiRepository) UpdateFriend(pFriend *models.Friend) error {

	if pFriend == nil {
		return errors.New("pFriend is nil")
	}

	where := models.Friend{UserID: pFriend.UserID, FriendUsername: pFriend.FriendUsername}
	// 同时更新多个字段
	result := s.db.Model(&models.Friend{}).Where(&where).Select("alias", "extend").Updates(pFriend)

	//updated records count
	s.logger.Debug("UpdateFriend result: ",
		zap.Int64("RowsAffected", result.RowsAffected),
		zap.Error(result.Error))

	if result.Error != nil {
		s.logger.Error("修改好友资料数据失败", zap.Error(result.Error))
		return result.Error
	}

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

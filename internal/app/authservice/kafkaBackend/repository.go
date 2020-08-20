package kafkaBackend

import (
	"github.com/lianmi/servers/internal/pkg/models"
	"go.uber.org/zap"
)
func (kc *KafkaClient) SaveAddFriend(pFriend *models.Friend) error {
	//使用事务同时更新好友数据
	tx := kc.GetTransaction()
	
	if err := tx.Save(pFriend).Error; err != nil {
		kc.logger.Error("更新好友表失败", zap.Error(err))
		tx.Rollback()

	}

	return nil
}
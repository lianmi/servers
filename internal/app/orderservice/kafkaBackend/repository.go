package kafkaBackend

import (
	"github.com/jinzhu/gorm"
	"github.com/lianmi/servers/internal/pkg/models"
	"go.uber.org/zap"
)

//GetTransaction 获取事务
func (kc *KafkaClient) GetTransaction() *gorm.DB {
	return kc.db.Begin()
}

//增加商品
func (kc *KafkaClient) SaveProduct(product *models.Product) error {
	//使用事务同时更新增加商品
	tx := kc.GetTransaction()

	if err := tx.Save(product).Error; err != nil {
		kc.logger.Error("更新Product表失败", zap.Error(err))
		tx.Rollback()
		return err
	}

	//提交
	tx.Commit()

	return nil
}

//删除商品
func (kc *KafkaClient) DeleteProduct(productID, username string) error {
	where := models.Product{ProductID: productID, Username: username}
	db := kc.db.Where(&where).Delete(models.Product{})
	err := db.Error
	if err != nil {
		kc.logger.Error("DeleteProduct", zap.Error(err))
		return err
	}
	count := db.RowsAffected
	kc.logger.Debug("DeleteProduct成功", zap.Int64("count", count))
	return nil
}

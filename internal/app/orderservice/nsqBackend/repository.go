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

//增加商品
func (nc *NsqClient) SaveProduct(product *models.Product) error {
	//使用事务同时更新增加商品
	tx := nc.GetTransaction()

	if err := tx.Save(product).Error; err != nil {
		nc.logger.Error("更新Product表失败", zap.Error(err))
		tx.Rollback()
		return err
	}

	//提交
	tx.Commit()

	return nil
}

//删除商品
func (nc *NsqClient) DeleteProduct(productID, username string) error {
	where := models.Product{ProductID: productID, Username: username}
	db := nc.db.Where(&where).Delete(models.Product{})
	err := db.Error
	if err != nil {
		nc.logger.Error("DeleteProduct", zap.Error(err))
		return err
	}
	count := db.RowsAffected
	nc.logger.Debug("DeleteProduct成功", zap.Int64("count", count))
	return nil
}

//保存商户的OPK, 批量
func (nc *NsqClient) SavePreKeys(prekeys []*models.Prekey) error {
	//使用事务批量保存商户的OPK
	tx := nc.GetTransaction()

	for _, prekey := range prekeys {
		if err := tx.Save(prekey).Error; err != nil {
			nc.logger.Error("保存prekey表失败", zap.Error(err))
			tx.Rollback()
			continue
		}
	}

	//提交
	tx.Commit()

	return nil
}


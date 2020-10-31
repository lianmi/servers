package repositories

import (
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	"github.com/lianmi/servers/internal/pkg/models"

	"go.uber.org/zap"
)

type OrderRepository interface {
	SaveProduct(product *models.Product) error
	DeleteProduct(productID, username string) error
	SavePreKeys(prekeys []*models.Prekey) error
}

type MysqlOrderRepository struct {
	logger    *zap.Logger
	db        *gorm.DB
	redisPool *redis.Pool
	base      *BaseRepository
}

func NewMysqlOrderRepository(logger *zap.Logger, db *gorm.DB, redisPool *redis.Pool) OrderRepository {
	return &MysqlOrderRepository{
		logger:    logger.With(zap.String("type", "OrderRepository")),
		db:        db,
		redisPool: redisPool,
		base:      NewBaseRepository(logger, db),
	}
}

//增加商品
func (m *MysqlOrderRepository) SaveProduct(product *models.Product) error {
	//使用事务同时更新增加商品
	tx := m.base.GetTransaction()

	if err := tx.Save(product).Error; err != nil {
		m.logger.Error("更新Product表失败", zap.Error(err))
		tx.Rollback()
		return err
	}

	//提交
	tx.Commit()

	return nil
}

//删除商品
func (m *MysqlOrderRepository) DeleteProduct(productID, username string) error {
	where := models.Product{ProductID: productID, Username: username}
	db := m.db.Where(&where).Delete(models.Product{})
	err := db.Error
	if err != nil {
		m.logger.Error("DeleteProduct", zap.Error(err))
		return err
	}
	count := db.RowsAffected
	m.logger.Debug("DeleteProduct成功", zap.Int64("count", count))
	return nil
}

//保存商户的OPK, 批量
func (m *MysqlOrderRepository) SavePreKeys(prekeys []*models.Prekey) error {
	//使用事务批量保存商户的OPK
	tx := m.base.GetTransaction()

	for _, prekey := range prekeys {
		if err := tx.Save(prekey).Error; err != nil {
			m.logger.Error("保存prekey表失败", zap.Error(err))
			tx.Rollback()
			continue
		}
	}

	//提交
	tx.Commit()

	return nil
}

package repositories

import (
	"github.com/gomodule/redigo/redis"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderRepository interface {
	AddProduct(product *models.Product) error
	UpdateProduct(product *models.Product) error
	DeleteProduct(productID, username string) error
	AddPreKeys(prekeys []*models.Prekey) error
	GetVipUserPrice(payType int) (*models.VipPrice, error)
	SaveChargeHistory(chargeHistory *models.ChargeHistory) error
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
func (m *MysqlOrderRepository) AddProduct(product *models.Product) error {

	if product == nil {
		return errors.New("product is nil")
	}
	//如果没有记录，则增加，如果有记录，则更新全部字段
	if err := m.db.Clauses(clause.OnConflict{DoNothing: true}).Create(product).Error; err != nil {
		m.logger.Error("增加Product失败", zap.Error(err))
		return err
	} else {
		m.logger.Debug("增加Product成功")
	}
	return nil

}

//修改商品
func (m *MysqlOrderRepository) UpdateProduct(product *models.Product) error {

	where := models.Product{
		ProductInfo: models.ProductInfo{
			ProductID: product.ProductInfo.ProductID,
		},
	}
	// 同时更新多个字段
	result := m.db.Model(&models.Product{}).Where(&where).Updates(product)

	//updated records count
	m.logger.Debug("UpdateProduct result: ",
		zap.Int64("RowsAffected", result.RowsAffected),
		zap.Error(result.Error))

	if result.Error != nil {
		m.logger.Error("UpdateProduct失败", zap.Error(result.Error))
		return result.Error
	}

	return nil
}

//删除商品
func (m *MysqlOrderRepository) DeleteProduct(productID, username string) error {
	where := models.Product{
		ProductInfo: models.ProductInfo{
			ProductID: productID,
			Username:  username,
		},
	}
	db2 := m.db.Where(&where).Delete(models.Product{})
	err := db2.Error
	if err != nil {
		m.logger.Error("DeleteProduct", zap.Error(err))
		return err
	}
	count := db2.RowsAffected
	m.logger.Debug("DeleteProduct成功", zap.Int64("count", count))
	return nil
}

//保存商户的OPK, 批量
func (m *MysqlOrderRepository) AddPreKeys(prekeys []*models.Prekey) error {

	for _, prekey := range prekeys {
		//如果没有记录，则增加，如果有记录，则更新全部字段
		if err := m.db.Clauses(clause.OnConflict{DoNothing: true}).Create(prekey).Error; err != nil {
			m.logger.Error("增加prekey失败", zap.Error(err))
			continue
		} else {
			m.logger.Debug("增加prekey成功")
		}
	}

	return nil
}

//根据 PayType获取价格信息
func (m *MysqlOrderRepository) GetVipUserPrice(payType int) (*models.VipPrice, error) {
	p := new(models.VipPrice)
	where := models.VipPrice{
		PayType: payType,
	}
	if err := m.db.Model(p).Where(&where).First(p).Error; err != nil {
		return nil, errors.Wrapf(err, "PayType not found[payType=%d]", payType)
	}
	return p, nil
}

func (m *MysqlOrderRepository) SaveChargeHistory(chargeHistory *models.ChargeHistory) error {
	return m.db.Save(chargeHistory).Error

}

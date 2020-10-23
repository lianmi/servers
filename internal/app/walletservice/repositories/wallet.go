package repositories

import (
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	// "github.com/lianmi/servers/internal/app/walletservice/nsqBackend"
	"github.com/lianmi/servers/internal/pkg/models"
	// "github.com/pkg/errors"
	"go.uber.org/zap"
)

type WalletRepository interface {
	SaveLnmcOrderTransferHistory(lnmcOrderTransferHistory *models.LnmcOrderTransferHistory) error
}

type MysqlWalletRepository struct {
	logger    *zap.Logger
	db        *gorm.DB
	redisPool *redis.Pool
	base      *BaseRepository
}

func NewMysqlWalletRepository(logger *zap.Logger, db *gorm.DB, redisPool *redis.Pool) WalletRepository {
	return &MysqlWalletRepository{
		logger:    logger.With(zap.String("type", "WalletRepository")),
		db:        db,
		redisPool: redisPool,
		base:      NewBaseRepository(logger, db),
	}
}

//数据库操作，将订单到账及退款记录到 MySQL
func (m *MysqlWalletRepository) SaveLnmcOrderTransferHistory(lnmcOrderTransferHistory *models.LnmcOrderTransferHistory) error {
	tx := m.base.GetTransaction()

	if err := tx.Save(lnmcOrderTransferHistory).Error; err != nil {
		m.logger.Error("更新订单到账及退款记录失败", zap.Error(err))
		tx.Rollback()
		return err

	}

	//提交
	tx.Commit()

	return nil

}

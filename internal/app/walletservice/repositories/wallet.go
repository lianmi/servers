package repositories

import (
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	// "github.com/lianmi/servers/internal/app/walletservice/nsqBackend"
	// "github.com/lianmi/servers/internal/pkg/models"
	// "github.com/pkg/errors"
	"go.uber.org/zap"
)

type WalletRepository interface {
}

type MysqlWalletRepository struct {
	logger    *zap.Logger
	db        *gorm.DB
	redisPool *redis.Pool
	// nsq     *nsqBackend.NsqClient
	base *BaseRepository
}

func NewMysqlWalletRepository(logger *zap.Logger, db *gorm.DB, redisPool *redis.Pool) WalletRepository {
	return &MysqlWalletRepository{
		logger:    logger.With(zap.String("type", "WalletRepository")),
		db:        db,
		redisPool: redisPool,
		// nsq:     nc,
		base: NewBaseRepository(logger, db),
	}
}

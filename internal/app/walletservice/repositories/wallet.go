package repositories

import (
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	// "github.com/lianmi/servers/internal/app/walletservice/kafkaBackend"
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
	// kafka     *kafkaBackend.KafkaClient
	base      *BaseRepository
}

func NewMysqlWalletRepository(logger *zap.Logger, db *gorm.DB, redisPool *redis.Pool) WalletRepository {
	return &MysqlWalletRepository{
		logger:    logger.With(zap.String("type", "WalletRepository")),
		db:        db,
		redisPool: redisPool,
		// kafka:     kc,
		base:      NewBaseRepository(logger, db),
	}
}


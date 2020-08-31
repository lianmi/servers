package repositories

import (
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	// "github.com/lianmi/servers/internal/pkg/models"

	// "github.com/lianmi/servers/internal/app/orderservice/kafkaBackend"
	// "github.com/pkg/errors"
	"go.uber.org/zap"
)

type OrderRepository interface {
}

type MysqlOrderRepository struct {
	logger    *zap.Logger
	db        *gorm.DB
	redisPool *redis.Pool
	// kafka     *kafkaBackend.KafkaClient
	base *BaseRepository
}

func NewMysqlOrderRepository(logger *zap.Logger, db *gorm.DB, redisPool *redis.Pool) OrderRepository {
	return &MysqlOrderRepository{
		logger:    logger.With(zap.String("type", "OrderRepository")),
		db:        db,
		redisPool: redisPool,
		// kafka:     kc,
		base: NewBaseRepository(logger, db),
	}
}

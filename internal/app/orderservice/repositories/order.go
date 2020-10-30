package repositories

import (
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	// "github.com/lianmi/servers/internal/pkg/models"

	// "github.com/lianmi/servers/internal/app/orderservice/nsqMq"
	// "github.com/pkg/errors"
	"go.uber.org/zap"
)

type OrderRepository interface {
}

type MysqlOrderRepository struct {
	logger    *zap.Logger
	db        *gorm.DB
	redisPool *redis.Pool
	// nsq     *nsqMq.NsqClient
	base *BaseRepository
}

func NewMysqlOrderRepository(logger *zap.Logger, db *gorm.DB, redisPool *redis.Pool) OrderRepository {
	return &MysqlOrderRepository{
		logger:    logger.With(zap.String("type", "OrderRepository")),
		db:        db,
		redisPool: redisPool,
		// nsq:     nc,
		base: NewBaseRepository(logger, db),
	}
}

package repositories

import (
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	// "github.com/lianmi/servers/internal/app/chatservice/kafkaBackend"
	// "github.com/lianmi/servers/internal/pkg/models"
	// "github.com/pkg/errors"
	"go.uber.org/zap"
)

type ChatRepository interface {
}

type MysqlChatRepository struct {
	logger    *zap.Logger
	db        *gorm.DB
	redisPool *redis.Pool
	// kafka     *kafkaBackend.KafkaClient
	base      *BaseRepository
}

func NewMysqlChatRepository(logger *zap.Logger, db *gorm.DB, redisPool *redis.Pool) ChatRepository {
	return &MysqlChatRepository{
		logger:    logger.With(zap.String("type", "ChatRepository")),
		db:        db,
		redisPool: redisPool,
		// kafka:     kc,
		base:      NewBaseRepository(logger, db),
	}
}


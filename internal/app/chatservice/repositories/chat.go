package repositories

import (
	"github.com/gomodule/redigo/redis"
	"gorm.io/gorm"
	// "github.com/lianmi/servers/internal/pkg/models"

	"go.uber.org/zap"
)

type ChatRepository interface {
}

type MysqlChatRepository struct {
	logger    *zap.Logger
	db        *gorm.DB
	redisPool *redis.Pool
	base      *BaseRepository
}

func NewMysqlChatRepository(logger *zap.Logger, db *gorm.DB, redisPool *redis.Pool) ChatRepository {
	return &MysqlChatRepository{
		logger:    logger.With(zap.String("type", "ChatRepository")),
		db:        db,
		redisPool: redisPool,
		base:      NewBaseRepository(logger, db),
	}
}

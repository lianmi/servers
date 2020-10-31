package services

import (
	// "context"
	// "fmt"
	// "time"

	"github.com/lianmi/servers/internal/app/chatservice/repositories"
	// LMCommon "github.com/lianmi/servers/internal/common"
	// "github.com/lianmi/servers/internal/pkg/models"
	// "github.com/pkg/errors"

	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
)

type ChatService interface {
}

type DefaultApisService struct {
	logger     *zap.Logger
	Repository repositories.MysqlChatRepository
	redisPool  *redis.Pool
}

func NewApisService(logger *zap.Logger, Repository repositories.MysqlChatRepository, redisPool *redis.Pool) ChatService {
	return &DefaultApisService{
		logger:     logger.With(zap.String("type", "ChatService")),
		Repository: Repository,
		redisPool:  redisPool,
	}
}

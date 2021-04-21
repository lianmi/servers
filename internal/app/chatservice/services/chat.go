package services

import (
	Order "github.com/lianmi/servers/api/proto/order"
	"github.com/lianmi/servers/internal/app/chatservice/repositories"

	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
)

type ChatService interface {
}

type DefaultApisService struct {
	logger             *zap.Logger
	Repository         repositories.ChatRepository
	redisPool          *redis.Pool
	orderGrpcClientSvc Order.LianmiOrderClient //order的grpc client
	// walletGrpcClientSvc Wallet.LianmiWalletClient //wallet的grpc client
}

func NewApisService(logger *zap.Logger, repository repositories.ChatRepository, redisPool *redis.Pool, oc Order.LianmiOrderClient) ChatService {
	return &DefaultApisService{
		logger:             logger.With(zap.String("type", "ChatService")),
		Repository:         repository,
		redisPool:          redisPool,
		orderGrpcClientSvc: oc,
		// walletGrpcClientSvc: wc,
	}
}

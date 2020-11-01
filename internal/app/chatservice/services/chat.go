package services

import (
	// "context"
	// "fmt"
	// "time"
	Auth "github.com/lianmi/servers/api/proto/auth"
	Order "github.com/lianmi/servers/api/proto/order"
	Wallet "github.com/lianmi/servers/api/proto/wallet"
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
	logger              *zap.Logger
	Repository          repositories.ChatRepository
	redisPool           *redis.Pool
	authGrpcClientSvc   Auth.LianmiAuthClient     //auth的grpc client
	orderGrpcClientSvc  Order.LianmiOrderClient   //order的grpc client
	walletGrpcClientSvc Wallet.LianmiWalletClient //wallet的grpc client
}

func NewApisService(logger *zap.Logger, repository repositories.ChatRepository, redisPool *redis.Pool, lc Auth.LianmiAuthClient, oc Order.LianmiOrderClient, wc Wallet.LianmiWalletClient) ChatService {
	return &DefaultApisService{
		logger:              logger.With(zap.String("type", "ChatService")),
		Repository:          repository,
		redisPool:           redisPool,
		authGrpcClientSvc:   lc,
		orderGrpcClientSvc:  oc,
		walletGrpcClientSvc: wc,
	}
}

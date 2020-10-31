package services

import (
	"context"
	// "fmt"
	// "time"

	// Order "github.com/lianmi/servers/api/proto/order"
	Wallet "github.com/lianmi/servers/api/proto/wallet"
	"github.com/lianmi/servers/internal/app/orderservice/repositories"
	// LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	// "github.com/pkg/errors"

	"github.com/gomodule/redigo/redis"
	"github.com/lianmi/servers/internal/pkg/blockchain"
	"go.uber.org/zap"
)

type OrderService interface {
	SaveProduct(product *models.Product) error
	DeleteProduct(productID, username string) error
	SavePreKeys(prekeys []*models.Prekey) error
	//订单完成或退款
	TransferByOrder(ctx context.Context, req *Wallet.TransferReq) (*Wallet.TransferResp, error)
}

type DefaultApisService struct {
	logger     *zap.Logger
	Repository repositories.OrderRepository
	redisPool  *redis.Pool
	walletSvc  Wallet.LianmiWalletClient
	ethService *blockchain.Service //连接以太坊geth的websocket
}

func NewApisService(logger *zap.Logger, Repository repositories.OrderRepository, redisPool *redis.Pool, walletSvc Wallet.LianmiWalletClient, ethService *blockchain.Service) OrderService {
	return &DefaultApisService{
		logger:     logger.With(zap.String("type", "OrderService")),
		Repository: Repository,
		redisPool:  redisPool,
		walletSvc:  walletSvc,
		ethService: ethService,
	}
}

func (s *DefaultApisService) TransferByOrder(ctx context.Context, req *Wallet.TransferReq) (*Wallet.TransferResp, error) {
	return s.walletSvc.TransferByOrder(ctx, req)
}

func (s *DefaultApisService) SaveProduct(product *models.Product) error {
	return s.Repository.SaveProduct(product)
}

func (s *DefaultApisService) DeleteProduct(productID, username string) error {
	return s.Repository.DeleteProduct(productID, username)
}

func (s *DefaultApisService) SavePreKeys(prekeys []*models.Prekey) error {
	return s.Repository.SavePreKeys(prekeys)
}

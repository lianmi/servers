package services

import (
	"github.com/gomodule/redigo/redis"
	// Wallet "github.com/lianmi/servers/api/proto/wallet"
	"github.com/lianmi/servers/internal/app/orderservice/repositories"
	"github.com/lianmi/servers/internal/pkg/models"
	"go.uber.org/zap"
)

type OrderService interface {
	AddProduct(product *models.Product) error
	UpdateProduct(product *models.Product) error
	DeleteProduct(productID, username string) error
	//订单完成或退款
	// TransferByOrder(ctx context.Context, req *Wallet.TransferReq) (*Wallet.TransferResp, error)

	//根据 PayType获取价格信息
	GetVipUserPrice(payType int) (*models.VipPrice, error)

	SaveChargeHistory(chargeHistory *models.ChargeHistory) error

	//获取彩票公证处的rsa公钥数据, 预先保存在redis里的
	GetNotaryServicePublickey(businessUsername string) (string, error)
}

type DefaultApisService struct {
	logger     *zap.Logger
	Repository repositories.OrderRepository
	redisPool  *redis.Pool
	// walletSvc  Wallet.LianmiWalletClient
}

func NewApisService(logger *zap.Logger, repository repositories.OrderRepository, redisPool *redis.Pool) OrderService {
	return &DefaultApisService{
		logger:     logger.With(zap.String("type", "OrderService")),
		Repository: repository,
		redisPool:  redisPool,
		// walletSvc:  walletSvc,
	}
}

// func (s *DefaultApisService) TransferByOrder(ctx context.Context, req *Wallet.TransferReq) (*Wallet.TransferResp, error) {
// 	return s.walletSvc.TransferByOrder(ctx, req)
// }

func (s *DefaultApisService) AddProduct(product *models.Product) error {
	return s.Repository.AddProduct(product)
}

func (s *DefaultApisService) UpdateProduct(product *models.Product) error {
	return s.Repository.UpdateProduct(product)
}

func (s *DefaultApisService) DeleteProduct(productID, username string) error {
	return s.Repository.DeleteProduct(productID, username)
}

//根据 PayType获取价格信息
func (s *DefaultApisService) GetVipUserPrice(payType int) (*models.VipPrice, error) {
	return s.Repository.GetVipUserPrice(payType)
}

func (s *DefaultApisService) SaveChargeHistory(chargeHistory *models.ChargeHistory) error {
	return s.Repository.SaveChargeHistory(chargeHistory)
}

//获取彩票公证处的rsa公钥数据, 预先保存在redis里的
func (s *DefaultApisService) GetNotaryServicePublickey(businessUsername string) (string, error) {
	return s.Repository.GetNotaryServicePublickey(businessUsername)
}

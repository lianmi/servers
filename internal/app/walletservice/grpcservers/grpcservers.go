/*
本文件实现grpc client的远程调用服务，在此实现对应的逻辑
*/
package grpcservers

import (
	"context"
	Wallet "github.com/lianmi/servers/api/proto/wallet"
	"github.com/lianmi/servers/internal/app/walletservice/services"
	"go.uber.org/zap"
)

type WalletServer struct {
	logger *zap.Logger

	service services.WalletService
}

func NewWalletServer(logger *zap.Logger, ps services.WalletService) (*WalletServer, error) {
	return &WalletServer{
		logger:  logger,
		service: ps,
	}, nil
}

//订单完成或退款
func (s *WalletServer) TransferByOrder(ctx context.Context, req *Wallet.TransferReq) (*Wallet.TransferResp, error) {
	return s.service.TransferByOrder(ctx, req)
}

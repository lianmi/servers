/*
本文件实现grpc client的远程调用服务，在此实现对应的逻辑
*/
package grpcservers

import (
	// "context"
	// Order "github.com/lianmi/servers/api/proto/order"
	"github.com/lianmi/servers/internal/app/orderservice/services"
	"go.uber.org/zap"
)

type OrderServer struct {
	logger *zap.Logger

	service services.OrderService
}

func NewOrderServer(logger *zap.Logger, ps services.OrderService) (*OrderServer, error) {
	return &OrderServer{
		logger:  logger,
		service: ps,
	}, nil
}

// //订单完成或退款
// func (s *OrderServer) TransferByOrder(ctx context.Context, req *Order.TransferReq) (*Order.TransferResp, error) {
// 	return s.service.TransferByOrder(ctx, req)
// }

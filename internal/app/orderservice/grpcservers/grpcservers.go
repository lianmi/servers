/*
本文件实现 api/proto/order/Service.proto 的全部Grpc接口
*/
package grpcservers

import (
	// "context"
	// Order "github.com/lianmi/servers/api/proto/order"
	"github.com/lianmi/servers/internal/app/orderservice/services"
	"go.uber.org/zap"
)

type OrderGrpcServer struct {
	logger *zap.Logger

	service services.OrderService
}

func NewOrderGrpcServer(logger *zap.Logger, ps services.OrderService) (*OrderGrpcServer, error) {
	return &OrderGrpcServer{
		logger:  logger,
		service: ps,
	}, nil
}

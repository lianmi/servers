/*
本文件实现grpc client的远程调用服务，在此实现对应的逻辑
*/
package grpcservers

import (
	"context"
	Service "github.com/lianmi/servers/api/proto/service"
	"github.com/lianmi/servers/internal/app/authservice/services"
	"go.uber.org/zap"
)

type AuthGrpcServer struct {
	logger *zap.Logger

	service services.AuthService
}

func NewAuthGrpcServer(logger *zap.Logger, ps services.AuthService) (*AuthGrpcServer, error) {
	return &AuthGrpcServer{
		logger:  logger,
		service: ps,
	}, nil
}

func (s *AuthGrpcServer) GetUser(ctx context.Context, in *Service.UserReq) (*Service.UserRsp, error) {
	return s.service.GetUser(ctx, in)
}

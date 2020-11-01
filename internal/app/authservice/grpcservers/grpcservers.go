/*
本文件实现 api/proto/auth/Service.proto 全部Grpc接口
*/
package grpcservers

import (
	"context"
	Auth "github.com/lianmi/servers/api/proto/auth"
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

func (s *AuthGrpcServer) GetUser(ctx context.Context, in *Auth.UserReq) (*Auth.UserRsp, error) {
	return s.service.GetUser(ctx, in)
}

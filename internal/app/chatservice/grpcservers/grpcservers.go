/*
本文件实现grpc client的远程调用服务，在此实现对应的逻辑
*/
package grpcservers

import (
	// "context"
	// Msg "github.com/lianmi/servers/api/proto/msg"
	"github.com/lianmi/servers/internal/app/chatservice/services"
	"go.uber.org/zap"
)

type ChatGrpcServer struct {
	logger *zap.Logger

	service services.ChatService
}

func NewChatGrpcServer(logger *zap.Logger, ps services.ChatService) (*ChatGrpcServer, error) {
	return &ChatGrpcServer{
		logger:  logger,
		service: ps,
	}, nil
}


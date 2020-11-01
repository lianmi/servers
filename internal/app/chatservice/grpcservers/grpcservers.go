/*
本文件实现 api/proto/msg/grpc.proto 全部Grpc接口
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

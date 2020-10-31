package grpcservers

import (
	"github.com/google/wire"
	Msg "github.com/lianmi/servers/api/proto/msg"
	"github.com/lianmi/servers/internal/pkg/transports/grpc"
	stdgrpc "google.golang.org/grpc"
)

func CreateInitServersFn(
	ps *ChatServer,
) grpc.InitServers {
	return func(s *stdgrpc.Server) {
		Msg.RegisterLianmiChatServer(s, ps)
	}
}

var ProviderSet = wire.NewSet(NewChatServer, CreateInitServersFn)

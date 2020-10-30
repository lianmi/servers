package grpcservers

import (
	"github.com/google/wire"
	Service "github.com/lianmi/servers/api/proto/service"
	"github.com/lianmi/servers/internal/pkg/transports/grpc"
	stdgrpc "google.golang.org/grpc"
)

func CreateInitServersFn(
	ps *AuthGrpcServer,
) grpc.InitServers {
	return func(s *stdgrpc.Server) {
		Service.RegisterLianmiApisServer(s, ps)
	}
}

var ProviderSet = wire.NewSet(NewAuthGrpcServer, CreateInitServersFn)

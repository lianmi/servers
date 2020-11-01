package grpcservers

import (
	"github.com/google/wire"
	Auth "github.com/lianmi/servers/api/proto/auth"
	"github.com/lianmi/servers/internal/pkg/transports/grpc"
	stdgrpc "google.golang.org/grpc"
)

func CreateInitServersFn(
	ps *AuthGrpcServer,
) grpc.InitServers {
	return func(s *stdgrpc.Server) {
		Auth.RegisterLianmiAuthServer(s, ps)
	}
}

var ProviderSet = wire.NewSet(NewAuthGrpcServer, CreateInitServersFn)

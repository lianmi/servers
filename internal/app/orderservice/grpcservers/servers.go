package grpcservers

import (
	"github.com/google/wire"
	Order "github.com/lianmi/servers/api/proto/order"
	"github.com/lianmi/servers/internal/pkg/transports/grpc"
	stdgrpc "google.golang.org/grpc"
)

func CreateInitServersFn(
	ps *OrderGrpcServer,
) grpc.InitServers {
	return func(s *stdgrpc.Server) {
		Order.RegisterLianmiOrderServer(s, ps)
	}
}

var ProviderSet = wire.NewSet(NewOrderGrpcServer, CreateInitServersFn)

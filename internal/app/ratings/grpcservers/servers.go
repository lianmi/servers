package grpcservers

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/api/proto"
	"github.com/lianmi/servers/internal/pkg/transports/grpc"
	stdgrpc "google.golang.org/grpc"
)



func CreateInitServersFn(
	ps *RatingsServer,
) grpc.InitServers {
	return func(s *stdgrpc.Server) {
		proto.RegisterRatingsServer(s, ps)
	}
}

var ProviderSet = wire.NewSet(NewRatingsServer, CreateInitServersFn)
package grpcservers

import (
	"github.com/google/wire"
	Wallet "github.com/lianmi/servers/api/proto/wallet"
	"github.com/lianmi/servers/internal/pkg/transports/grpc"
	stdgrpc "google.golang.org/grpc"
)

func CreateInitServersFn(
	ps *WalletServer,
) grpc.InitServers {
	return func(s *stdgrpc.Server) {
		Wallet.RegisterLianmiWalletServer(s, ps)
	}
}

var ProviderSet = wire.NewSet(NewWalletServer, CreateInitServersFn)

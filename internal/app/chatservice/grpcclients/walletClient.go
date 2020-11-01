package grpcclients

import (
	Wallet "github.com/lianmi/servers/api/proto/wallet"
	"github.com/lianmi/servers/internal/pkg/transports/grpc"
	"github.com/pkg/errors"
)

func NewWalletClient(client *grpc.Client) (Wallet.LianmiWalletClient, error) {
	conn, err := client.Dial("cloud.lianmi.im.wallet.LianmiWallet")
	if err != nil {
		return nil, errors.Wrap(err, "wallet grpc client dial error")
	} else {

	}
	c := Wallet.NewLianmiWalletClient(conn)

	return c, nil
}

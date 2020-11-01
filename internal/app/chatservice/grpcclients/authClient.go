package grpcclients

import (
	Auth "github.com/lianmi/servers/api/proto/auth"
	"github.com/lianmi/servers/internal/pkg/transports/grpc"
	"github.com/pkg/errors"
)

func NewApisClient(client *grpc.Client) (Auth.LianmiAuthClient, error) {
	conn, err := client.Dial("cloud.lianmi.im.service.LianmiAuth")
	if err != nil {
		return nil, errors.Wrap(err, "service grpc client dial error")
	} else {

	}
	c := Auth.NewLianmiAuthClient(conn)

	return c, nil
}

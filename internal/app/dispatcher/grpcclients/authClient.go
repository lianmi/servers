package grpcclients

import (
	Service "github.com/lianmi/servers/api/proto/service"
	"github.com/lianmi/servers/internal/pkg/transports/grpc"
	"github.com/pkg/errors"
)

func NewApisClient(client *grpc.Client) (Service.LianmiApisClient, error) {
	conn, err := client.Dial("cloud.lianmi.im.service.LianmiApis")
	if err != nil {
		return nil, errors.Wrap(err, "service grpc client dial error")
	} else {

	}
	c := Service.NewLianmiApisClient(conn)

	return c, nil
}

package grpcclients

import (
	Order "github.com/lianmi/servers/api/proto/order"
	"github.com/lianmi/servers/internal/pkg/transports/grpc"
	"github.com/pkg/errors"
)

func NewOrderClient(client *grpc.Client) (Order.LianmiOrderClient, error) {
	conn, err := client.Dial("cloud.lianmi.im.order.LianmiOrder")
	if err != nil {
		return nil, errors.Wrap(err, "order grpc client dial error")
	} else {

	}
	c := Order.NewLianmiOrderClient(conn)

	return c, nil
}

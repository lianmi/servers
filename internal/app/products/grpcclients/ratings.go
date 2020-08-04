package grpcclients

import (
	"github.com/pkg/errors"
	"github.com/lianmi/servers/api/proto"
	"github.com/lianmi/servers/internal/pkg/transports/grpc"
)

func NewRatingsClient(client *grpc.Client) (proto.RatingsClient, error) {
	conn, err := client.Dial("Ratings")
	if err != nil {
		return nil, errors.Wrap(err, "detail client dial error")
	}
	c := proto.NewRatingsClient(conn)

	return c, nil
}

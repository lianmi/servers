// +build wireinject

package services

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/app/orderservice/grpcclients"
	"github.com/lianmi/servers/internal/app/orderservice/repositories"
	"github.com/lianmi/servers/internal/pkg/config"
	"github.com/lianmi/servers/internal/pkg/consul"
	"github.com/lianmi/servers/internal/pkg/database"
	"github.com/lianmi/servers/internal/pkg/jaeger"
	"github.com/lianmi/servers/internal/pkg/log"
	"github.com/lianmi/servers/internal/pkg/redis"
	"github.com/lianmi/servers/internal/pkg/transports/grpc"
)

var testProviderSet = wire.NewSet(
	log.ProviderSet,
	config.ProviderSet,
	database.ProviderSet,
	redis.ProviderSet,
	consul.ProviderSet,
	grpcclients.ProviderSet,
	grpc.ProviderSet,
	jaeger.ProviderSet,
	ProviderSet,
)

func CreateApisService(cf string, sto repositories.OrderRepository) (OrderService, error) {
	panic(wire.Build(testProviderSet))
}

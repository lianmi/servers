// +build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/lianmi/servers/internal/app/orderservice"
	// "github.com/lianmi/servers/internal/app/orderservice/grpcclients"
	"github.com/lianmi/servers/internal/app/orderservice/grpcservers"
	"github.com/lianmi/servers/internal/app/orderservice/nsqMq"
	"github.com/lianmi/servers/internal/app/orderservice/repositories"
	"github.com/lianmi/servers/internal/app/orderservice/services"
	"github.com/lianmi/servers/internal/pkg/app"
	"github.com/lianmi/servers/internal/pkg/config"
	"github.com/lianmi/servers/internal/pkg/consul"
	"github.com/lianmi/servers/internal/pkg/database"
	"github.com/lianmi/servers/internal/pkg/jaeger"
	"github.com/lianmi/servers/internal/pkg/log"
	"github.com/lianmi/servers/internal/pkg/redis"
	"github.com/lianmi/servers/internal/pkg/transports/grpc"
)

var providerSet = wire.NewSet(
	log.ProviderSet,
	config.ProviderSet,
	database.ProviderSet,
	services.ProviderSet,
	repositories.ProviderSet,
	consul.ProviderSet,
	jaeger.ProviderSet,
	redis.ProviderSet,
	grpc.ProviderSet,
	grpcservers.ProviderSet,
	nsqMq.ProviderSet,
	orderservice.ProviderSet,
	// grpcclients.ProviderSet,
)

func CreateApp(cf string) (*app.Application, error) {
	panic(wire.Build(providerSet))
}

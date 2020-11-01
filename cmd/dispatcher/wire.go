// +build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/lianmi/servers/internal/pkg/channel"
	"github.com/lianmi/servers/internal/app/dispatcher"
	"github.com/lianmi/servers/internal/app/dispatcher/controllers"
	"github.com/lianmi/servers/internal/app/dispatcher/grpcclients"
	"github.com/lianmi/servers/internal/app/dispatcher/nsqMq"
	"github.com/lianmi/servers/internal/app/dispatcher/repositories"
	"github.com/lianmi/servers/internal/app/dispatcher/services"
	"github.com/lianmi/servers/internal/pkg/app"
	"github.com/lianmi/servers/internal/pkg/config"
	"github.com/lianmi/servers/internal/pkg/consul"
	"github.com/lianmi/servers/internal/pkg/database"
	"github.com/lianmi/servers/internal/pkg/jaeger"
	"github.com/lianmi/servers/internal/pkg/log"
	"github.com/lianmi/servers/internal/pkg/redis"
	"github.com/lianmi/servers/internal/pkg/transports/grpc"
	"github.com/lianmi/servers/internal/pkg/transports/http"
	"github.com/lianmi/servers/internal/pkg/transports/mqtt"
)

var providerSet = wire.NewSet(
	log.ProviderSet,
	config.ProviderSet,
	database.ProviderSet,
	services.ProviderSet,
	repositories.ProviderSet,
	consul.ProviderSet,
	http.ProviderSet,
	redis.ProviderSet,
	jaeger.ProviderSet,
	channel.ProviderSet,
	nsqMq.ProviderSet,
	mqtt.ProviderSet,
	dispatcher.ProviderSet,
	controllers.ProviderSet,
	grpc.ProviderSet,
	grpcclients.ProviderSet,
)

func CreateApp(cf string) (*app.Application, error) {
	panic(wire.Build(providerSet))
}

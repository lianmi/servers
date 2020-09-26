// +build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/lianmi/servers/internal/app/walletservice"
	// "github.com/lianmi/servers/internal/app/walletservice/controllers"
	"github.com/lianmi/servers/internal/app/walletservice/repositories"
	// "github.com/lianmi/servers/internal/app/walletservice/services"
	"github.com/lianmi/servers/internal/app/walletservice/nsqBackend"
	"github.com/lianmi/servers/internal/pkg/app"
	"github.com/lianmi/servers/internal/pkg/config"
	"github.com/lianmi/servers/internal/pkg/consul"
	"github.com/lianmi/servers/internal/pkg/database"
	"github.com/lianmi/servers/internal/pkg/jaeger"
	"github.com/lianmi/servers/internal/pkg/log"
	"github.com/lianmi/servers/internal/pkg/redis"
	// "github.com/lianmi/servers/internal/pkg/transports/http"
)

var providerSet = wire.NewSet(
	log.ProviderSet,
	config.ProviderSet,
	database.ProviderSet,
	// services.ProviderSet,
	repositories.ProviderSet,
	consul.ProviderSet,
	jaeger.ProviderSet,
	redis.ProviderSet,
	// http.ProviderSet,
	nsqBackend.ProviderSet,
	walletservice.ProviderSet,
	// controllers.ProviderSet,
)

func CreateApp(cf string) (*app.Application, error) {
	panic(wire.Build(providerSet))
}

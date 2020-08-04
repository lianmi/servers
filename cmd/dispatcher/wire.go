// +build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/lianmi/servers/internal/pkg/app"
	"github.com/lianmi/servers/internal/pkg/config"
	"github.com/lianmi/servers/internal/pkg/log"
	"github.com/lianmi/servers/internal/app/channel"
	"github.com/lianmi/servers/internal/pkg/transports/kafka"
	"github.com/lianmi/servers/internal/pkg/transports/mqtt"
	"github.com/lianmi/servers/internal/app/dispatcher"
	"github.com/lianmi/servers/internal/pkg/redis"

)

var providerSet = wire.NewSet(
	log.ProviderSet,
	config.ProviderSet,
	redis.ProviderSet,
	channel.ProviderSet,
	kafka.ProviderSet,
	mqtt.ProviderSet,
	dispatcher.ProviderSet,

)

func CreateApp(cf string) (*app.Application, error) {
	panic(wire.Build(providerSet))
}

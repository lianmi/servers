// +build wireinject

package services

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/app/dispatcher/grpcclients"
	"github.com/lianmi/servers/internal/app/dispatcher/repositories"
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

func CreateLianmiApisService(cf string, sto repositories.LianmiRepository) (LianmiApisService, error) {
	panic(wire.Build(testProviderSet))
}

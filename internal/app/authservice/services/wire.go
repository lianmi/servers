// +build wireinject

package services

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/pkg/config"
	"github.com/lianmi/servers/internal/pkg/database"
	"github.com/lianmi/servers/internal/pkg/redis"
	"github.com/lianmi/servers/internal/pkg/log"
	"github.com/lianmi/servers/internal/app/authservice/repositories"
)

var testProviderSet = wire.NewSet(
	log.ProviderSet,
	config.ProviderSet,
	database.ProviderSet,
	redis.ProviderSet,
	ProviderSet,
)

func CreateLianmiApisService(cf string, sto repositories.LianmiRepository) (LianmiApisService, error) {
	panic(wire.Build(testProviderSet))
}

// +build wireinject

package controllers

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/app/authservice/repositories"
	"github.com/lianmi/servers/internal/app/authservice/services"
	"github.com/lianmi/servers/internal/pkg/config"
	"github.com/lianmi/servers/internal/pkg/database"
	"github.com/lianmi/servers/internal/pkg/log"
	"github.com/lianmi/servers/internal/pkg/redis"
)

var testProviderSet = wire.NewSet(
	log.ProviderSet,
	config.ProviderSet,
	database.ProviderSet,
	services.ProviderSet,
	redis.ProviderSet,
	// repositories.ProviderSet, //不需要！！！
	ProviderSet,
)

func CreateLianmiApisController(cf string, sto repositories.LianmiRepository) (*LianmiApisController, error) {
	panic(wire.Build(testProviderSet))
}

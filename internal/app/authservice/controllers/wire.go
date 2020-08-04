// +build wireinject

package controllers

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/pkg/config"
	"github.com/lianmi/servers/internal/pkg/database"
	"github.com/lianmi/servers/internal/pkg/log"
	"github.com/lianmi/servers/internal/app/authservice/services"
	"github.com/lianmi/servers/internal/app/authservice/repositories"
)

var testProviderSet = wire.NewSet(
	log.ProviderSet,
	config.ProviderSet,
	database.ProviderSet,
	services.ProviderSet,
	// repositories.ProviderSet, //不需要！！！
	ProviderSet,
)


func CreateUsersController(cf string, sto repositories.UsersRepository) (*UsersController, error) {
	panic(wire.Build(testProviderSet))
}

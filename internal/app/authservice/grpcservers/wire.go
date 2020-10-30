// +build wireinject

package grpcservers

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/app/authservice/services"
	"github.com/lianmi/servers/internal/pkg/config"
	"github.com/lianmi/servers/internal/pkg/database"
	"github.com/lianmi/servers/internal/pkg/log"
)

var testProviderSet = wire.NewSet(
	log.ProviderSet,
	config.ProviderSet,
	database.ProviderSet,
	ProviderSet,
)

func CreateAuthGrpcServer(cf string, service services.AuthService) (*AuthGrpcServer, error) {
	panic(wire.Build(testProviderSet))
}

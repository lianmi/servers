// +build wireinject

package repositories

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/pkg/config"
	"github.com/lianmi/servers/internal/pkg/database"
	"github.com/lianmi/servers/internal/app/walletservice/nsqBackend"
	"github.com/lianmi/servers/internal/pkg/redis"
	"github.com/lianmi/servers/internal/pkg/log"
)

var testProviderSet = wire.NewSet(
	log.ProviderSet,
	config.ProviderSet,
	database.ProviderSet,
	redis.ProviderSet,
	nsqBackend.ProviderSet,
	ProviderSet,
)

func CreateWalletRepository(f string) (WalletRepository, error) {
	panic(wire.Build(testProviderSet))
}


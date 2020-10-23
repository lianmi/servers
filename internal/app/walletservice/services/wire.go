// +build wireinject

package services

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/app/walletservice/repositories"
	"github.com/lianmi/servers/internal/pkg/blockchain"
	"github.com/lianmi/servers/internal/pkg/config"
	"github.com/lianmi/servers/internal/pkg/database"
	"github.com/lianmi/servers/internal/pkg/log"
	"github.com/lianmi/servers/internal/pkg/redis"
)

var testProviderSet = wire.NewSet(
	log.ProviderSet,
	config.ProviderSet,
	database.ProviderSet,
	redis.ProviderSet,
	blockchain.ProviderSet,
	ProviderSet,
)

func CreateApisService(cf string, sto repositories.WalletRepository) (WalletService, error) {
	panic(wire.Build(testProviderSet))
}

//

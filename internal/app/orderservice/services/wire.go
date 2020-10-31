// +build wireinject

package services

import (
	"github.com/google/wire"
	Wallet "github.com/lianmi/servers/api/proto/wallet"
	"github.com/lianmi/servers/internal/app/orderservice/repositories"
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

func CreateApisService(cf string, sto repositories.OrderRepository, wlc Wallet.LianmiWalletClient) (OrderService, error) {
	panic(wire.Build(testProviderSet))
}

//

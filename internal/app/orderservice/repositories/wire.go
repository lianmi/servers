// +build wireinject

package repositories

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/app/orderservice/nsqMq"
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
	nsqMq.ProviderSet,
	ProviderSet,
)

func CreateOrderRepository(f string) (OrderRepository, error) {
	panic(wire.Build(testProviderSet))
}

// +build wireinject

package repositories

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/app/channel"
	"github.com/lianmi/servers/internal/app/dispatcher/nsqMq"
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
	channel.ProviderSet,
	ProviderSet,
)

func CreateLianmirRepository(f string) (LianmiRepository, error) {
	panic(wire.Build(testProviderSet))
}

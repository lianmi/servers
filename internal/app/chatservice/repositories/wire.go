// +build wireinject

package repositories

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/pkg/config"
	"github.com/lianmi/servers/internal/pkg/database"
	"github.com/lianmi/servers/internal/app/chatservice/kafkaBackend"
	"github.com/lianmi/servers/internal/pkg/redis"
	"github.com/lianmi/servers/internal/pkg/log"
)



var testProviderSet = wire.NewSet(
	log.ProviderSet,
	config.ProviderSet,
	database.ProviderSet,
	redis.ProviderSet,
	kafkaBackend.ProviderSet,
	ProviderSet,
)

func CreateChatRepository(f string) (ChatRepository, error) {
	panic(wire.Build(testProviderSet))
}


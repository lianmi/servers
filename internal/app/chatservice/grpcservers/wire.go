// +build wireinject

package grpcservers

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/app/chatservice/services"
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

func CreateChatServer(cf string, service services.ChatService) (*ChatServer, error) {
	panic(wire.Build(testProviderSet))
}

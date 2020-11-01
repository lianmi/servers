// +build wireinject

package channel

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/pkg/config"
	"github.com/lianmi/servers/internal/pkg/log"
	
)

var testProviderSet = wire.NewSet(
	log.ProviderSet,
	config.ProviderSet,
	channel.ProviderSet,
	ProviderSet,
)


func CreateChannel() (*channel.NsqMqttChannel, error) {
	panic(wire.Build(testProviderSet))
}

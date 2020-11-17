// +build wireinject

package multichannel

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/pkg/config"
	"github.com/lianmi/servers/internal/pkg/log"
)

var testProviderSet = wire.NewSet(
	log.ProviderSet,
	config.ProviderSet,
	multichannel.ProviderSet,
	ProviderSet,
)

func CreateChannel() (*multichannel.NsqChannel, error) {
	panic(wire.Build(testProviderSet))
}

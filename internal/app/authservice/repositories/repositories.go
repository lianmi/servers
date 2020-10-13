package repositories

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewMysqlLianmiRepository)
//var MockProviderSet = wire.NewSet(wire.InterfaceValue(new(LianmiRepository),new(MockLianmiRepository)))

package repositories

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewMysqlChatRepository)
//var MockProviderSet = wire.NewSet(wire.InterfaceValue(new(LianmiRepository),new(MockLianmiRepository)))

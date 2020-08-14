package repositories

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewMysqlWalletRepository)

//var MockProviderSet = wire.NewSet(wire.InterfaceValue(new(UsersRepository),new(MockUsersRepository)))

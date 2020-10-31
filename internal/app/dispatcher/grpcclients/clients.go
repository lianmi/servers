package grpcclients

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewApisClient, NewWalletClient, NewOrderClient)

// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package grpcservers

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/app/walletservice/services"
	"github.com/lianmi/servers/internal/pkg/config"
	"github.com/lianmi/servers/internal/pkg/database"
	"github.com/lianmi/servers/internal/pkg/log"
)

// Injectors from wire.go:

func CreateWalletServer(cf string, service services.WalletService) (*WalletServer, error) {
	viper, err := config.New(cf)
	if err != nil {
		return nil, err
	}
	options, err := log.NewOptions(viper)
	if err != nil {
		return nil, err
	}
	logger, err := log.New(options)
	if err != nil {
		return nil, err
	}
	walletServer, err := NewWalletServer(logger, service)
	if err != nil {
		return nil, err
	}
	return walletServer, nil
}

// wire.go:

var testProviderSet = wire.NewSet(log.ProviderSet, config.ProviderSet, database.ProviderSet, ProviderSet)
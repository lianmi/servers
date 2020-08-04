// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package grpcservers

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/app/details/services"
	"github.com/lianmi/servers/internal/pkg/config"
	"github.com/lianmi/servers/internal/pkg/database"
	"github.com/lianmi/servers/internal/pkg/log"
)

// Injectors from wire.go:

func CreateDetailsServer(cf string, service services.DetailsService) (*DetailsServer, error) {
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
	detailsServer, err := NewDetailsServer(logger, service)
	if err != nil {
		return nil, err
	}
	return detailsServer, nil
}

// wire.go:

var testProviderSet = wire.NewSet(log.ProviderSet, config.ProviderSet, database.ProviderSet, ProviderSet)

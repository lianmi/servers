// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package services

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/app/dispatcher/grpcclients"
	"github.com/lianmi/servers/internal/app/dispatcher/repositories"
	"github.com/lianmi/servers/internal/pkg/config"
	"github.com/lianmi/servers/internal/pkg/consul"
	"github.com/lianmi/servers/internal/pkg/database"
	"github.com/lianmi/servers/internal/pkg/jaeger"
	"github.com/lianmi/servers/internal/pkg/log"
	"github.com/lianmi/servers/internal/pkg/redis"
	"github.com/lianmi/servers/internal/pkg/transports/grpc"
)

// Injectors from wire.go:

func CreateLianmiApisService(cf string, sto repositories.LianmiRepository) (LianmiApisService, error) {
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
	consulOptions, err := consul.NewOptions(viper)
	if err != nil {
		return nil, err
	}
	configuration, err := jaeger.NewConfiguration(viper, logger)
	if err != nil {
		return nil, err
	}
	tracer, err := jaeger.New(configuration)
	if err != nil {
		return nil, err
	}
	clientOptions, err := grpc.NewClientOptions(viper, tracer)
	if err != nil {
		return nil, err
	}
	client, err := grpc.NewClient(consulOptions, clientOptions)
	if err != nil {
		return nil, err
	}
	lianmiAuthClient, err := grpcclients.NewAuthClient(client)
	if err != nil {
		return nil, err
	}
	lianmiOrderClient, err := grpcclients.NewOrderClient(client)
	if err != nil {
		return nil, err
	}
	lianmiWalletClient, err := grpcclients.NewWalletClient(client)
	if err != nil {
		return nil, err
	}
	lianmiApisService := NewLianmiApisService(logger, sto, lianmiAuthClient, lianmiOrderClient, lianmiWalletClient)
	return lianmiApisService, nil
}

// wire.go:

var testProviderSet = wire.NewSet(log.ProviderSet, config.ProviderSet, database.ProviderSet, redis.ProviderSet, consul.ProviderSet, grpcclients.ProviderSet, grpc.ProviderSet, jaeger.ProviderSet, ProviderSet)

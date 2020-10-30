// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package main

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/app/walletservice"
	"github.com/lianmi/servers/internal/app/walletservice/grpcservers"
	"github.com/lianmi/servers/internal/app/walletservice/nsqMq"
	"github.com/lianmi/servers/internal/app/walletservice/repositories"
	"github.com/lianmi/servers/internal/app/walletservice/services"
	"github.com/lianmi/servers/internal/pkg/app"
	"github.com/lianmi/servers/internal/pkg/blockchain"
	"github.com/lianmi/servers/internal/pkg/config"
	"github.com/lianmi/servers/internal/pkg/consul"
	"github.com/lianmi/servers/internal/pkg/database"
	"github.com/lianmi/servers/internal/pkg/jaeger"
	"github.com/lianmi/servers/internal/pkg/log"
	"github.com/lianmi/servers/internal/pkg/redis"
	"github.com/lianmi/servers/internal/pkg/transports/grpc"
)

// Injectors from wire.go:

func CreateApp(cf string) (*app.Application, error) {
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
	walletserviceOptions, err := walletservice.NewOptions(viper, logger)
	if err != nil {
		return nil, err
	}
	nsqOptions, err := nsqMq.NewNsqOptions(viper)
	if err != nil {
		return nil, err
	}
	databaseOptions, err := database.NewOptions(viper, logger)
	if err != nil {
		return nil, err
	}
	db, err := database.New(databaseOptions)
	if err != nil {
		return nil, err
	}
	redisOptions, err := redis.NewRedisOptions(viper, logger)
	if err != nil {
		return nil, err
	}
	pool, err := redis.New(redisOptions)
	if err != nil {
		return nil, err
	}
	walletRepository := repositories.NewMysqlWalletRepository(logger, db, pool)
	blockchainOptions, err := blockchain.NewEthClientProviderOptions(viper, logger)
	if err != nil {
		return nil, err
	}
	service, err := blockchain.New(blockchainOptions, logger)
	if err != nil {
		return nil, err
	}
	nsqClient := nsqMq.NewNsqClient(nsqOptions, walletRepository, pool, logger, service)
	serverOptions, err := grpc.NewServerOptions(viper)
	if err != nil {
		return nil, err
	}
	walletService := services.NewApisService(logger, walletRepository, pool, service)
	walletServer, err := grpcservers.NewWalletServer(logger, walletService)
	if err != nil {
		return nil, err
	}
	initServers := grpcservers.CreateInitServersFn(walletServer)
	consulOptions, err := consul.NewOptions(viper)
	if err != nil {
		return nil, err
	}
	client, err := consul.New(consulOptions)
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
	server, err := grpc.NewServer(serverOptions, logger, initServers, client, tracer)
	if err != nil {
		return nil, err
	}
	application, err := walletservice.NewApp(walletserviceOptions, logger, nsqClient, service, server)
	if err != nil {
		return nil, err
	}
	return application, nil
}

// wire.go:

var providerSet = wire.NewSet(log.ProviderSet, config.ProviderSet, database.ProviderSet, services.ProviderSet, repositories.ProviderSet, consul.ProviderSet, jaeger.ProviderSet, redis.ProviderSet, grpc.ProviderSet, grpcservers.ProviderSet, nsqMq.ProviderSet, walletservice.ProviderSet, blockchain.ProviderSet)

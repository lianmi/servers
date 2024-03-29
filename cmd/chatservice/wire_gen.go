// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package main

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/app/chatservice"
	"github.com/lianmi/servers/internal/app/chatservice/grpcclients"
	"github.com/lianmi/servers/internal/app/chatservice/grpcservers"
	"github.com/lianmi/servers/internal/app/chatservice/nsqMq"
	"github.com/lianmi/servers/internal/app/chatservice/repositories"
	"github.com/lianmi/servers/internal/app/chatservice/services"
	"github.com/lianmi/servers/internal/pkg/app"
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
	chatserviceOptions, err := chatservice.NewOptions(viper, logger)
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
	nsqClient := nsqMq.NewNsqClient(nsqOptions, db, pool, logger)
	serverOptions, err := grpc.NewServerOptions(viper)
	if err != nil {
		return nil, err
	}
	chatRepository := repositories.NewMysqlChatRepository(logger, db, pool)
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
	lianmiOrderClient, err := grpcclients.NewOrderClient(client)
	if err != nil {
		return nil, err
	}
	chatService := services.NewApisService(logger, chatRepository, pool, lianmiOrderClient)
	chatGrpcServer, err := grpcservers.NewChatGrpcServer(logger, chatService)
	if err != nil {
		return nil, err
	}
	initServers := grpcservers.CreateInitServersFn(chatGrpcServer)
	apiClient, err := consul.New(consulOptions)
	if err != nil {
		return nil, err
	}
	server, err := grpc.NewServer(serverOptions, logger, initServers, apiClient, tracer)
	if err != nil {
		return nil, err
	}
	application, err := chatservice.NewApp(chatserviceOptions, logger, nsqClient, server)
	if err != nil {
		return nil, err
	}
	return application, nil
}

// wire.go:

var providerSet = wire.NewSet(log.ProviderSet, config.ProviderSet, database.ProviderSet, services.ProviderSet, repositories.ProviderSet, consul.ProviderSet, jaeger.ProviderSet, redis.ProviderSet, grpc.ProviderSet, grpcservers.ProviderSet, nsqMq.ProviderSet, chatservice.ProviderSet, grpcclients.ProviderSet)

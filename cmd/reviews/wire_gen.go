// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package main

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/app/reviews"
	"github.com/lianmi/servers/internal/app/reviews/controllers"
	"github.com/lianmi/servers/internal/app/reviews/grpcservers"
	"github.com/lianmi/servers/internal/app/reviews/repositories"
	"github.com/lianmi/servers/internal/app/reviews/services"
	"github.com/lianmi/servers/internal/pkg/app"
	"github.com/lianmi/servers/internal/pkg/config"
	"github.com/lianmi/servers/internal/pkg/consul"
	"github.com/lianmi/servers/internal/pkg/database"
	"github.com/lianmi/servers/internal/pkg/jaeger"
	"github.com/lianmi/servers/internal/pkg/log"
	"github.com/lianmi/servers/internal/pkg/transports/grpc"
	"github.com/lianmi/servers/internal/pkg/transports/http"
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
	reviewsOptions, err := reviews.NewOptions(viper, logger)
	if err != nil {
		return nil, err
	}
	httpOptions, err := http.NewOptions(viper)
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
	reviewsRepository := repositories.NewMysqlReviewsRepository(logger, db)
	reviewsService := services.NewReviewService(logger, reviewsRepository)
	reviewsController := controllers.NewReviewsController(logger, reviewsService)
	initControllers := controllers.CreateInitControllersFn(reviewsController)
	configuration, err := jaeger.NewConfiguration(viper, logger)
	if err != nil {
		return nil, err
	}
	tracer, err := jaeger.New(configuration)
	if err != nil {
		return nil, err
	}
	engine := http.NewRouter(httpOptions, logger, initControllers, tracer)
	consulOptions, err := consul.NewOptions(viper)
	if err != nil {
		return nil, err
	}
	client, err := consul.New(consulOptions)
	if err != nil {
		return nil, err
	}
	server, err := http.New(httpOptions, logger, engine, client)
	if err != nil {
		return nil, err
	}
	serverOptions, err := grpc.NewServerOptions(viper)
	if err != nil {
		return nil, err
	}
	reviewsServer, err := grpcservers.NewReviewsServer(logger, reviewsService)
	if err != nil {
		return nil, err
	}
	initServers := grpcservers.CreateInitServersFn(reviewsServer)
	grpcServer, err := grpc.NewServer(serverOptions, logger, initServers, client, tracer)
	if err != nil {
		return nil, err
	}
	application, err := reviews.NewApp(reviewsOptions, logger, server, grpcServer)
	if err != nil {
		return nil, err
	}
	return application, nil
}

// wire.go:

var providerSet = wire.NewSet(log.ProviderSet, config.ProviderSet, database.ProviderSet, services.ProviderSet, consul.ProviderSet, jaeger.ProviderSet, http.ProviderSet, grpc.ProviderSet, reviews.ProviderSet, repositories.ProviderSet, controllers.ProviderSet, grpcservers.ProviderSet)

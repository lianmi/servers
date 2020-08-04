// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package services

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/app/ratings/repositories"
	"github.com/lianmi/servers/internal/pkg/config"
	"github.com/lianmi/servers/internal/pkg/database"
	"github.com/lianmi/servers/internal/pkg/log"
)

// Injectors from wire.go:

func CreateRatingsService(cf string, sto repositories.RatingsRepository) (RatingsService, error) {
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
	ratingsService := NewRatingService(logger, sto)
	return ratingsService, nil
}

// wire.go:

var testProviderSet = wire.NewSet(log.ProviderSet, config.ProviderSet, database.ProviderSet, ProviderSet)

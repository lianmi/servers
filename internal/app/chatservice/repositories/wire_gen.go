// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package repositories

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/app/chatservice/kafkaBackend"
	"github.com/lianmi/servers/internal/pkg/config"
	"github.com/lianmi/servers/internal/pkg/database"
	"github.com/lianmi/servers/internal/pkg/log"
	"github.com/lianmi/servers/internal/pkg/redis"
)

// Injectors from wire.go:

func CreateChatRepository(f string) (ChatRepository, error) {
	viper, err := config.New(f)
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
	chatRepository := NewMysqlChatRepository(logger, db, pool)
	return chatRepository, nil
}

// wire.go:

var testProviderSet = wire.NewSet(log.ProviderSet, config.ProviderSet, database.ProviderSet, redis.ProviderSet, kafkaBackend.ProviderSet, ProviderSet)

package app

import (
	"subscription-service/internal/app/grpcApp"
	"subscription-service/internal/config"
	"subscription-service/internal/repository/cachedDbRepository"
	"subscription-service/internal/repository/postgresRepository"
	"subscription-service/internal/repository/redisRepository"
	"subscription-service/internal/service"
	"subscription-service/pkg/logger"
)

func New(config *config.Config, logger logger.Logger) *grpcApp.App {
	pool, err := postgresRepository.NewPool(config)
	if err != nil {
		panic(err)
	}

	redisClient, err := redisRepository.NewRedisClient(*config)
	if err != nil {
		panic(err)
	}

	dbRepo := postgresRepository.NewPostgresRepository(pool)
	redis := redisRepository.NewRedisRepository(redisClient)
	cachedDbRepo := cachedDbRepository.NewCachedPostgresRepository(dbRepo, redis)
	subService := service.NewSubscriptionService(cachedDbRepo)

	return grpcApp.NewApp(subService, config.GrpcPort, cachedDbRepo, logger)
}

package app

import (
	"subscription-service/internal/app/grpcApp"
	"subscription-service/internal/config"
	"subscription-service/internal/repository/postgresRepository"
	"subscription-service/internal/service"
)

func New(config *config.Config) *grpcApp.App {
	pool, err := postgresRepository.NewPool(config)
	if err != nil {
		panic(err)
	}
	repo := postgresRepository.NewPostgresRepository(pool)
	subService := service.NewSubscriptionService(repo)

	return grpcApp.NewApp(subService, config.GrpcPort)
}

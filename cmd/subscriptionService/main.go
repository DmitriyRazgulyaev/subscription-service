package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"subscription-service/internal/app"
	"subscription-service/internal/config"
	"subscription-service/pkg/logger"
	"syscall"
	"time"
)

const (
	stopTimeout = time.Second * 10
)

func main() {
	envPath := os.Getenv("ENV_PATH")
	if envPath == "" {
		envPath = "./config/.env"
	}

	cfg, err := config.ParseConfigFromEnv(envPath)
	if err != nil {
		panic(err)
	}

	lo, err := logger.NewLogger(cfg.Environment)
	if err != nil {
		panic(err)
	}

	application := app.New(cfg, *lo)

	ctx := context.Background()

	log.Printf("starting server at :%d", cfg.GrpcPort)
	go application.MustRun()
	go application.RunRest(ctx, cfg.GrpcPort, cfg.GatewayPort)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	stopCtx, cancel := context.WithTimeout(ctx, stopTimeout)
	defer cancel()
	application.GracefulStop(stopCtx)
}

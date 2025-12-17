package main

import (
	"log"
	"os"
	"os/signal"
	"subscription-service/internal/app"
	"subscription-service/internal/config"
	"subscription-service/pkg/logger"
	"syscall"
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

	log.Printf("starting server at :%d", cfg.GrpcPort)
	go application.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	application.GracefulStop()
}

package main

import (
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"os"
	"subscription-service/internal/config"
)

func main() {
	var migrationsPath string
	var command string

	flag.StringVar(&migrationsPath, "path", "./migrations", "path to migrations directory")
	flag.StringVar(&command, "command", "up", "migration command: up, down, version")
	flag.Parse()

	envPath := os.Getenv("ENV_PATH")
	if envPath == "" {
		envPath = "./config/.env"
	}

	cfg, err := config.ParseConfigFromEnv(envPath)
	if err != nil {
		panic(err)
	}

	url := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable",
		cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresDb)

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		url,
	)
	if err != nil {
		panic(fmt.Errorf("failed to create migrate instance: %v", err))
	}
	defer m.Close()

	switch command {
	case "up":
		if err := m.Up(); err != nil {
			panic(fmt.Errorf("failed to apply migrations: %v", err))
		}
		fmt.Println("Migrations applied successfully!")

	case "down":
		if err := m.Down(); err != nil {
			panic(fmt.Errorf("failed to rollback migrations: %v", err))
		}
		fmt.Println("Migrations rolled back successfully!")

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			panic(fmt.Errorf("failed to get version: %v", err))
		}
		fmt.Printf("Current version: %d, Dirty: %t\n", version, dirty)

	default:
		fmt.Printf("unknown command: %s\n", command)
		fmt.Println("Available commands: up, down, version")
		os.Exit(1)
	}
}

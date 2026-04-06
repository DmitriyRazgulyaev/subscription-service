package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"time"
)

type Config struct {
	Environment string `env:"ENV" env-default:"development"`
	GrpcPort    int    `env:"GRPC_PORT" env-default:"50051"`
	GatewayPort int    `env:"GATEWAY_PORT" env-default:"8080"`

	PostgresVersion  int    `env:"POSTGRES_VERSION" env-default:"15"`
	PostgresDb       string `env:"POSTGRES_DB" env-default:"postgres"`
	PostgresUser     string `env:"POSTGRES_USER" env-default:"postgres"`
	PostgresPassword string `env:"POSTGRES_PASSWORD" env-default:"postgres"`
	PostgresHost     string `env:"POSTGRES_HOST" env-default:"db"`
	PostgresPort     int    `env:"POSTGRES_PORT" env-default:"5432"`

	RedisAddr        string        `env:"REDIS_ADDR" env-default:"redis:6379"`
	RedisMaxRetries  int           `env:"REDIS_MAX_RETRIES" env-default:"5"`
	RedisDialTimeout time.Duration `env:"REDIS_DIAL_TIMEOUT" env-default:"10000000000"`
	RedisTimeout     time.Duration `env:"REDIS_TIMEOUT" env-default:"5000000000"`
}

func ParseConfigFromEnv(envPath string) (*Config, error) {
	cfg := &Config{}
	if err := cleanenv.ReadConfig(envPath, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

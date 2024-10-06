package config

import (
	"github.com/caarlos0/env/v9"
	"github.com/joho/godotenv"
	"log"
	"time"
)

type Config struct {
	// Environment: local, dev, test, prod
	Environment string `env:"ENVIRONMENT" envDefault:"local"`
	// REST API server
	Address     string        `env:"ADDRESS" envDefault:"localhost:8080"`
	Timeout     time.Duration `env:"TIMEOUT" envDefault:"5s"`
	IdleTimeout time.Duration `env:"IDLE_TIMEOUT" envDefault:"60s"`
	// Postgres connection
	PgHost     string `env:"PG_HOST" envDefault:"localhost"`
	PgPort     string `env:"PG_PORT" envDefault:"5432"`
	PgUser     string `env:"PG_USER" envDefault:"postgres"`
	PgPass     string `env:"PG_PASS" envDefault:"postgres"`
	PgDatabase string `env:"PG_DATABASE" envDefault:"songs"`
}

func MustLoad() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Unable to load .env file: %e", err)
	}

	cfg := Config{}
	err = env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Unable to parse ennvironment variables: %e", err)
	}

	return &cfg
}

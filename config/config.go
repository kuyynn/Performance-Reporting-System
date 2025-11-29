package config

import (
	"log"

	"github.com/caarlos0/env/v9"
)

type Config struct {
	AppPort     string `env:"APP_PORT" envDefault:"8080"`
	PostgresDSN string `env:"POSTGRES_DSN"`
	MongoURI    string `env:"MONGO_URI" envDefault:"mongodb://localhost:27017"`
	MongoDB     string `env:"MONGO_DB" envDefault:"prestasi_db"`
	JWTSecret   string `env:"JWT_SECRET" envDefault:"changeme"`
}

func LoadConfig() *Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to load env: %v", err)
	}

	if cfg.PostgresDSN == "" {
		log.Println("[WARN] POSTGRES_DSN is empty, Postgres may not connect")
	}

	return &cfg
}

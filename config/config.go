package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/caarlos0/env/v9"
)

type Config struct {
	AppPort     string `env:"APP_PORT" envDefault:"3000"`
	PostgresDSN string `env:"POSTGRES_DSN"`
	MongoURI    string `env:"MONGO_URI" envDefault:"mongodb://localhost:27017"`
	MongoDB     string `env:"MONGO_DB" envDefault:"uas_db"`
	JWTSecret   string `env:"JWT_SECRET" envDefault:"changeme"`
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("WARNING: .env file not found or failed to load")
	}

	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed load env: %v", err)
	}

	return &cfg
}

package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Env            string
	Port           string
	LogLevel       string
	APIKey         string
	DatabaseDriver string
	DatabaseDSN    string
}

var AppConfig *Config

func Load() {
	_ = godotenv.Load()

	AppConfig = &Config{
		Env:            getEnv("APP_ENV", "dev"),
		Port:           getEnv("PORT", "8080"),
		LogLevel:       getEnv("LOG_LEVEL", "debug"),
		APIKey:         os.Getenv("API_KEY"),
		DatabaseDriver: getEnv("DB_DRIVER", "sqlite"),
		DatabaseDSN:    getEnv("DB_DSN", "file:tmp/app.db?_foreign_keys=1"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

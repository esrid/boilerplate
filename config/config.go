// Package config provides configurations for the application.
package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSslmode  string
}

func New() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file")
		return nil, err
	}

	cfg := &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBSslmode:  os.Getenv("DB_SSLMODE"),
	}

	return cfg, nil
}

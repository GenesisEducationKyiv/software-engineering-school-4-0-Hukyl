package config

import (
	"log/slog"
	"os"
)

type Config struct {
	DatabaseService string
	DatabaseDSN     string
}

func NewFromEnv() Config {
	databaseService := os.Getenv("DATABASE_SERVICE")
	if databaseService == "" {
		slog.Error("DATABASE_SERVICE is not set")
	}
	databaseDSN := os.Getenv("DATABASE_DSN")
	if databaseDSN == "" {
		slog.Error("DATABASE_DSN is not set")
	}
	return Config{
		DatabaseService: databaseService,
		DatabaseDSN:     databaseDSN,
	}
}

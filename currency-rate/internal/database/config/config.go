package config

import (
	"log/slog"
	"os"
)

type Config struct {
	DatabaseService string
	DatabaseDSN     string
}

func getOrError(key string) string {
	value := os.Getenv(key)
	if value == "" {
		slog.Error("value is not set", slog.Any("key", key))
	}
	return value
}

func NewFromEnv() Config {
	return Config{
		DatabaseService: getOrError("DATABASE_SERVICE"),
		DatabaseDSN:     getOrError("DATABASE_DSN"),
	}
}

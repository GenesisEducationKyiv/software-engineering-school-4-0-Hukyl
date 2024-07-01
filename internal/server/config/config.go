package config

import (
	"log/slog"
	"os"
)

type Config struct {
	Port string
}

func NewFromEnv() Config {
	port := os.Getenv("PORT")
	if port == "" {
		slog.Error("PORT is not set")
	}
	return Config{
		Port: port,
	}
}

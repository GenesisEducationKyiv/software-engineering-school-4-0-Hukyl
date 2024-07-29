package config

import (
	"log/slog"
	"os"
)

type Config struct {
	BrokerURI string
	QueueName string
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
		BrokerURI: getOrError("BROKER_URI"),
		QueueName: getOrError("QUEUE_NAME"),
	}
}

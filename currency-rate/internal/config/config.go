package config

import (
	"log/slog"
	"os"
)

type Config struct {
	CurrencyBeaconAPIKey string
	RateRefreshCropSpec  string
	RateQueueName        string
	UserQueueName        string
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
		CurrencyBeaconAPIKey: getOrError("CURRENCY_BEACON_API_KEY"),
		RateRefreshCropSpec:  getOrError("RATE_REFRESH_CRON_SPEC"),
		RateQueueName:        getOrError("RATE_QUEUE_NAME"),
		UserQueueName:        getOrError("USER_QUEUE_NAME"),
	}
}

package config

import (
	"log/slog"
	"os"
)

type Config struct {
	Debug                   bool
	NotificationCropSpec    string
	VictoriaMetricsPushURL  string
	RateQueueName           string
	UserQueueName           string
	UserCompensateQueueName string
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
		Debug:                   getOrError("DEBUG") == "true",
		NotificationCropSpec:    getOrError("NOTIFICATION_CRON_SPEC"),
		VictoriaMetricsPushURL:  getOrError("VICTORIA_METRICS_PUSH_URL"),
		RateQueueName:           getOrError("RATE_QUEUE_NAME"),
		UserQueueName:           getOrError("USER_QUEUE_NAME"),
		UserCompensateQueueName: getOrError("USER_COMPENSATE_QUEUE_NAME"),
	}
}

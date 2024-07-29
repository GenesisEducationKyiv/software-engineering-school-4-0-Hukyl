package config

import (
	"log/slog"
	"os"
)

type Config struct {
	FromEmail    string
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
}

func getOrError(key string) string {
	value := os.Getenv(key)
	if value == "" {
		slog.Error("env value is not set", slog.Any("key", key))
	}
	return value
}

func NewFromEnv() Config {
	return Config{
		FromEmail:    getOrError("EMAIL_FROM"),
		SMTPHost:     getOrError("SMTP_HOST"),
		SMTPPort:     getOrError("SMTP_PORT"),
		SMTPUser:     getOrError("SMTP_USER"),
		SMTPPassword: getOrError("SMTP_PASSWORD"),
	}
}

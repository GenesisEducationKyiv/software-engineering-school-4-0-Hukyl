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

func NewFromEnv() Config {
	fromEmail := os.Getenv("EMAIL_FROM")
	if fromEmail == "" {
		slog.Error("EMAIL_FROM is not set")
	}
	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost == "" {
		slog.Error("SMTP_HOST is not set")
	}
	smtpPort := os.Getenv("SMTP_PORT")
	if smtpPort == "" {
		slog.Error("SMTP_PORT is not set")
	}
	smtpUser := os.Getenv("SMTP_USER")
	if smtpUser == "" {
		slog.Error("SMTP_USER is not set")
	}
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	if smtpPassword == "" {
		slog.Error("SMTP_PASSWORD is not set")
	}
	return Config{
		FromEmail:    fromEmail,
		SMTPHost:     smtpHost,
		SMTPPort:     smtpPort,
		SMTPUser:     smtpUser,
		SMTPPassword: smtpPassword,
	}
}

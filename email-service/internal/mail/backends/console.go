package backends

import (
	"context"
	"log/slog"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/mail/config"
)

var logger *slog.Logger

func getLogger() *slog.Logger {
	if logger == nil {
		logger = slog.Default().With(slog.String("src", "mailer"))
	}
	return logger
}

type ConsoleMailer struct {
	config config.Config
}

func (cm *ConsoleMailer) SendEmail(
	_ context.Context, emails []string,
	subject, message string,
) error {
	getLogger().Info(
		"sending email",
		slog.Any("fromEmail", cm.config.FromEmail),
		slog.Any("toEmails", emails),
		slog.Any("subject", subject),
		slog.Any("message", message),
	)
	return nil
}

func NewConsoleMailer(config config.Config) *ConsoleMailer {
	return &ConsoleMailer{config: config}
}

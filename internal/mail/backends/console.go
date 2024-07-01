package backends

import (
	"context"
	"log/slog"

	"github.com/Hukyl/genesis-kma-school-entry/internal/mail/config"
)

type ConsoleMailer struct {
	config config.Config
}

func (cm *ConsoleMailer) SendEmail(
	_ context.Context, emails []string,
	subject, message string,
) error {
	slog.Info(
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

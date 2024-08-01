package mail

import (
	"context"
	"fmt"
	"log/slog"
)

var logger *slog.Logger

func getLogger() *slog.Logger {
	if logger == nil {
		logger = slog.Default().With(slog.String("src", "mailClient"))
	}
	return logger
}

type Mailer interface {
	SendEmail(ctx context.Context, emails []string, subject, message string) error
}

type Client struct {
	backend Mailer
}

func NewClient(backend Mailer) *Client {
	return &Client{backend: backend}
}

func (mc *Client) SendEmail(ctx context.Context, emails []string, subject, message string) error {
	getLogger().Info("sending mail", slog.Any("backend", mc.backend))
	err := mc.backend.SendEmail(ctx, emails, subject, message)
	if err != nil {
		return fmt.Errorf("email client: %w", err)
	}
	getLogger().Debug("mail sent")
	return nil
}

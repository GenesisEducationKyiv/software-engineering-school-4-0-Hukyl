package mail

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/Hukyl/genesis-kma-school-entry/mail/config"
	"github.com/Hukyl/genesis-kma-school-entry/settings"
	"github.com/go-gomail/gomail"
)

type Client struct {
	Config config.Config
}

func (mc *Client) SendEmail(ctx context.Context, email, message string) error {
	debug, ok := ctx.Value(settings.DebugKey).(bool)
	if ok && debug {
		return mc.SendEmailStub(email, message)
	}
	return mc.SendSMTPEmail(email, message)
}

func (mc *Client) SendEmailStub(email, message string) error {
	slog.Info(
		"sending email",
		slog.Any("fromEmail", mc.Config.FromEmail),
		slog.Any("toEmail", email),
		slog.Any("message", message),
	)
	return nil
}

func (mc *Client) SendSMTPEmail(email, message string) error {
	config := mc.Config
	mail := gomail.NewMessage()
	mail.SetHeader("From", config.FromEmail)
	mail.SetHeader("To", email)
	mail.SetHeader("Subject", "USD-UAH exchange rate")
	mail.SetBody("text/html", message)

	port, _ := strconv.Atoi(config.SMTPPort)
	dialer := gomail.NewDialer(
		config.SMTPHost,
		port,
		config.SMTPUser,
		config.SMTPPassword,
	)
	if err := dialer.DialAndSend(mail); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

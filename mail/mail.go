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

func (mc *Client) SendEmail(ctx context.Context, email, subject, message string) error {
	debug, ok := ctx.Value(settings.DebugKey).(bool)
	if ok && debug {
		return mc.SendEmailStub(email, subject, message)
	}
	return mc.SendSMTPEmail(email, subject, message)
}

func (mc *Client) SendEmailStub(email, subject, message string) error {
	slog.Info(
		"sending email",
		slog.Any("fromEmail", mc.Config.FromEmail),
		slog.Any("toEmail", email),
		slog.Any("subject", subject),
		slog.Any("message", message),
	)
	return nil
}

func (mc *Client) SendSMTPEmail(email, subject, message string) error {
	config := mc.Config
	mail := gomail.NewMessage()
	mail.SetHeader("From", config.FromEmail)
	mail.SetHeader("To", email)
	mail.SetHeader("Subject", subject)
	mail.SetBody("text/html", message)

	port, err := strconv.Atoi(config.SMTPPort)
	if err != nil {
		return fmt.Errorf("failed to convert SMTP port to int: %w", err)
	}
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

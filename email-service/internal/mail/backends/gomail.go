package backends

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/mail/config"
	"github.com/VictoriaMetrics/metrics"
	"github.com/go-gomail/gomail"
)

var gomailOutgoingEmailsCounter = metrics.GetOrCreateCounter(
	`email_outgoing_total{backend="gomail"}`,
)
var gomailErrorsCounter = metrics.GetOrCreateCounter(
	`email_errors_total{backend="gomail"}`,
)

type GomailMailer struct {
	config config.Config
}

func (gm *GomailMailer) SendEmail(
	ctx context.Context, emails []string,
	subject, message string,
) error {
	mail := gomail.NewMessage()
	mail.SetHeader("From", gm.config.FromEmail)
	if len(emails) == 0 {
		getLogger().Error("no email recipients")
		return errors.New("no email recipients")
	}
	mail.SetHeader("To", emails[0])
	mail.SetHeader("Bcc", emails[1:]...)
	mail.SetHeader("Subject", subject)
	mail.SetBody("text/html", message)

	port, err := strconv.Atoi(gm.config.SMTPPort)
	if err != nil {
		getLogger().Error("parsing SMTP port", slog.Any("error", err))
		return fmt.Errorf("failed to convert SMTP port to int: %w", err)
	}
	dialer := gomail.NewDialer(
		gm.config.SMTPHost,
		port,
		gm.config.SMTPUser,
		gm.config.SMTPPassword,
	)

	done := make(chan error)
	go func() {
		done <- dialer.DialAndSend(mail)
	}()

	select {
	case <-ctx.Done():
		err := ctx.Err()
		gomailErrorsCounter.Inc()
		getLogger().Error("context done", slog.Any("error", err))
		return fmt.Errorf("email sending cancelled: %w", err)
	case err := <-done:
		if err != nil {
			gomailErrorsCounter.Inc()
			getLogger().Error("sending email", slog.Any("error", err))
			return fmt.Errorf("failed to send email: %w", err)
		}
		gomailOutgoingEmailsCounter.Inc()
		getLogger().Debug("email sent")
	}
	return nil
}

func NewGomailMailer(config config.Config) *GomailMailer {
	return &GomailMailer{config: config}
}

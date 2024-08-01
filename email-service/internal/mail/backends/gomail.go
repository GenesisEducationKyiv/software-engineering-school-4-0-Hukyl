package backends

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/mail/config"
	"github.com/go-gomail/gomail"
)

type GomailMailer struct {
	config config.Config
}

func (gm *GomailMailer) SendEmail(
	ctx context.Context, emails []string,
	subject, message string,
) error {
	if len(emails) == 0 {
		return errors.New("no email recipients")
	}

	mail := gomail.NewMessage()
	mail.SetHeader("From", gm.config.FromEmail)
	mail.SetHeader("To", emails[0])
	mail.SetHeader("Bcc", emails[1:]...)
	mail.SetHeader("Subject", subject)
	mail.SetBody("text/html", message)

	port, err := strconv.Atoi(gm.config.SMTPPort)
	if err != nil {
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
		return fmt.Errorf("email sending cancelled: %w", ctx.Err())
	case err := <-done:
		if err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}
	}
	return nil
}

func NewGomailMailer(config config.Config) *GomailMailer {
	return &GomailMailer{config: config}
}

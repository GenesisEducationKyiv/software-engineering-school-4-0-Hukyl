package backends

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/Hukyl/genesis-kma-school-entry/internal/mail/config"
	"github.com/go-gomail/gomail"
)

type GomailMailer struct {
	config config.Config
}

func (gm *GomailMailer) SendEmail(
	_ context.Context, emails []string,
	subject, message string,
) error {
	mail := gomail.NewMessage()
	mail.SetHeader("From", gm.config.FromEmail)
	if len(emails) == 0 {
		return errors.New("no email recipients")
	}
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
	if err := dialer.DialAndSend(mail); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

func NewGomailMailer(config config.Config) *GomailMailer {
	return &GomailMailer{config: config}
}

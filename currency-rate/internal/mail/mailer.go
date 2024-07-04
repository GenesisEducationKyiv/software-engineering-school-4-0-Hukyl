package mail

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/mail/config"
)

type MailData struct {
	Emails  []string `json:"emails"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
}

type MailerFacade struct {
	producer *Producer
}

func (m *MailerFacade) SendEmail(ctx context.Context, emails []string, subject string, message string) error {
	msg := MailData{
		Emails:  emails,
		Subject: subject,
		Body:    message,
	}
	slog.Info("sending email", slog.Any("userCount", len(emails)))
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshalling email message: %w", err)
	}
	if err := m.producer.Produce(ctx, msgBytes); err != nil {
		return fmt.Errorf("producing email message: %w", err)
	}
	return nil
}

func NewMailerFacade(config config.Config) (*MailerFacade, error) {
	producer, err := NewProducer(config)
	if err != nil {
		return nil, fmt.Errorf("creating producer: %w", err)
	}
	return &MailerFacade{producer}, nil
}

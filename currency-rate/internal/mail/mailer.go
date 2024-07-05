package mail

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/mail/transport"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/mail/transport/config"
)

var lastCommandID int

const eventType = "SendEmail"

type Command struct {
	ID        string `json:"commandID"`
	Type      string `json:"commandType"`
	Timestamp string `json:"timestamp"`
	Data      any    `json:"data"`
}

type mailData struct {
	Emails  []string `json:"emails"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
}

type MailerFacade struct {
	producer *transport.Producer
}

func (m *MailerFacade) createCommand(data any) Command {
	lastCommandID++
	return Command{
		ID:        strconv.Itoa(lastCommandID),
		Type:      eventType,
		Timestamp: time.Now().Format(time.RFC3339),
		Data:      data,
	}
}

func (m *MailerFacade) marshal(data Command) ([]byte, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshalling command: %w", err)
	}
	return bytes, nil
}

func (m *MailerFacade) SendEmail(
	ctx context.Context,
	emails []string, subject string, message string,
) error {
	data := m.createCommand(mailData{
		Emails:  emails,
		Subject: subject,
		Body:    message,
	})
	slog.Info("sending email", slog.Any("userCount", len(emails)))
	msgBytes, err := m.marshal(data)
	if err != nil {
		return err
	}
	if err := m.producer.Produce(ctx, msgBytes); err != nil {
		return fmt.Errorf("producing email message: %w", err)
	}
	return nil
}

func (m *MailerFacade) Close() error {
	return m.producer.Close()
}

func NewMailerFacade(config config.Config) (*MailerFacade, error) {
	producer, err := transport.NewProducer(config)
	if err != nil {
		return nil, fmt.Errorf("creating producer: %w", err)
	}
	return &MailerFacade{producer}, nil
}

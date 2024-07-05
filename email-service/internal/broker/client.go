package broker

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/broker/transport"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/broker/transport/config"
)

var mailTimeout = 5 * time.Second

const eventType = "SendEmail"

type MailData struct {
	Emails  []string `json:"emails"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
}

type Command struct {
	ID        string   `json:"commandID"`
	Type      string   `json:"commandType"`
	Timestamp string   `json:"timestamp"`
	Data      MailData `json:"data"`
}

type MailSender func(ctx context.Context, emails []string, subject, body string) error

type Client struct {
	consumer   *transport.Consumer
	stopSignal chan struct{}
}

func (c *Client) Subscribe(f MailSender) error {
	c.consumer.Subscribe(func(b []byte) error {
		ctx, cancel := context.WithTimeout(context.Background(), mailTimeout)
		defer cancel()
		command, err := c.unmarshal(b)
		if err != nil {
			return err
		}
		if command.Type != eventType {
			return nil
		}
		return f(ctx, command.Data.Emails, command.Data.Subject, command.Data.Body)
	})
	return nil
}

func (c *Client) unmarshal(data []byte) (*Command, error) {
	command := &Command{}
	if err := json.Unmarshal(data, command); err != nil {
		return nil, err
	}
	return command, nil
}

func (c *Client) Close() error {
	close(c.stopSignal)
	return c.consumer.Close()
}

func NewClient(config config.Config) (*Client, error) {
	consumer, err := transport.NewConsumer(config)
	if err != nil {
		slog.Error("creating consumer", slog.Any("error", err))
		return nil, err
	}
	stopSignal := make(chan struct{})
	go consumer.Listen(stopSignal)
	return &Client{consumer: consumer, stopSignal: stopSignal}, nil
}

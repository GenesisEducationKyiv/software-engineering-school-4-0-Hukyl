package broker

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/broker/config"
)

var mailTimeout = 5 * time.Second

type MailData struct {
	Emails  []string `json:"emails"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
}

type MailListeners func(ctx context.Context, emails []string, subject, body string) error

type Client struct {
	consumer   *Consumer
	stopSignal chan struct{}
}

func (c *Client) Subscribe(f MailListeners) error {
	c.consumer.listeners = append(c.consumer.listeners, func(b []byte) error {
		ctx, cancel := context.WithTimeout(context.Background(), mailTimeout)
		defer cancel()
		mailData, err := c.Unmarshal(b)
		if err != nil {
			return err
		}
		return f(ctx, mailData.Emails, mailData.Subject, mailData.Body)
	})
	return nil
}

func (c *Client) Unmarshal(data []byte) (*MailData, error) {
	mailData := &MailData{}
	if err := json.Unmarshal(data, mailData); err != nil {
		return nil, err
	}
	return mailData, nil
}

func (c *Client) Close() error {
	close(c.stopSignal)
	return c.consumer.Close()
}

func NewClient(config config.Config) (*Client, error) {
	consumer, err := NewConsumer(config)
	if err != nil {
		slog.Error("creating consumer", slog.Any("error", err))
		return nil, err
	}
	stopSignal := make(chan struct{})
	go consumer.Listen(stopSignal)
	return &Client{consumer: consumer, stopSignal: stopSignal}, nil
}

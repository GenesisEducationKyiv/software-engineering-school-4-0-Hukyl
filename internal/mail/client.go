package mail

import (
	"context"
	"fmt"
)

type Mailer interface {
	SendEmail(ctx context.Context, email, subject, message string) error
}

type Client struct {
	backend Mailer
}

func NewClient(backend Mailer) *Client {
	return &Client{backend: backend}
}

func (mc *Client) SendEmail(ctx context.Context, email, subject, message string) error {
	err := mc.backend.SendEmail(ctx, email, subject, message)
	if err != nil {
		return fmt.Errorf("email client: %w", err)
	}
	return nil
}

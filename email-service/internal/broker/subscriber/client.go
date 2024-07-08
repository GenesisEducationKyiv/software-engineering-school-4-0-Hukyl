package subscriber

import (
	"context"
	"encoding/json"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker/transport"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker/transport/config"
)

var subTimeout = 5 * time.Second

type Handler func(ctx context.Context, email string) error

const (
	subscribedEventType   = "Subscribe"
	unsubscribedEventType = "Unsubscribe"
)

type SubscribeEvent struct {
	broker.Event
	Data struct {
		Email string `json:"email"`
	} `json:"data"`
}

type Client struct {
	consumer   *transport.Consumer
	stopSignal chan struct{}
}

func (c *Client) handleWithEvent(eventName string, f Handler) func([]byte) error {
	return func(b []byte) error {
		ctx, cancel := context.WithTimeout(context.Background(), subTimeout)
		defer cancel()
		event, err := c.unmarshal(b)
		if err != nil {
			return err
		}
		if event.Event.Type != eventName {
			return nil
		}
		return f(ctx, event.Data.Email)
	}
}

func (c *Client) SubscribeCreate(f Handler) error {
	c.consumer.Subscribe(c.handleWithEvent(subscribedEventType, f))
	return nil
}

func (c *Client) SubscribeDelete(f Handler) error {
	c.consumer.Subscribe(c.handleWithEvent(unsubscribedEventType, f))
	return nil
}

func (c *Client) unmarshal(data []byte) (*SubscribeEvent, error) {
	event := &SubscribeEvent{}
	if err := json.Unmarshal(data, event); err != nil {
		return nil, err
	}
	return event, nil
}

func (c *Client) Close() error {
	close(c.stopSignal)
	return c.consumer.Close()
}

func NewClient(config config.Config) (*Client, error) {
	consumer, err := transport.NewConsumer(config)
	if err != nil {
		return nil, err
	}
	stopSignal := make(chan struct{})
	go consumer.Listen(stopSignal)
	return &Client{consumer: consumer, stopSignal: stopSignal}, nil
}
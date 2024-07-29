package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker/transport"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker/transport/config"
	"github.com/VictoriaMetrics/metrics"
)

var logger *slog.Logger

func getTotalReceivedUserMessagesCounter(eventName string) *metrics.Counter {
	// ? is it better to rework message delivering
	// ? logic to decrease the number of received messages?
	return metrics.GetOrCreateCounter(fmt.Sprintf(
		`broker_received_messages_total{consumer="subscriber_consumer", event="%s"}`, eventName,
	))
}

func getLogger() *slog.Logger {
	if logger == nil {
		logger = slog.Default().With(slog.Any("src", "subscriberClient"))
	}
	return logger
}

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
	getLogger().Debug(
		"subscribing to subscriber event",
		slog.Any("handler", f), slog.Any("eventName", eventName),
	)
	return func(b []byte) error {
		getTotalReceivedUserMessagesCounter(eventName).Inc()
		getLogger().Debug("received message", slog.Any("handler", f))
		ctx, cancel := context.WithTimeout(context.Background(), subTimeout)
		defer cancel()
		event, err := c.unmarshal(b)
		if err != nil {
			return err
		}
		if event.Event.Type != eventName {
			return nil
		}
		slog.Info(
			"delivering message",
			slog.Any("listener", f),
			slog.Any("eventName", event.Event.Type),
		)
		return f(ctx, event.Data.Email)
	}
}

func (c *Client) subscribeCreate(f Handler) error {
	c.consumer.Subscribe(c.handleWithEvent(subscribedEventType, f))
	return nil
}

func (c *Client) subscribeDelete(f Handler) error {
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
	getLogger().Info("closing subscriber client")
	close(c.stopSignal)
	return c.consumer.Close()
}

func NewClient(config config.Config, subscribe Handler, unsubscribe Handler) (*Client, error) {
	consumer, err := transport.NewConsumer(config)
	if err != nil {
		return nil, err
	}
	stopSignal := make(chan struct{})

	client := &Client{consumer: consumer, stopSignal: stopSignal}

	if err = client.subscribeCreate(subscribe); err != nil {
		return nil, err
	}
	if err = client.subscribeDelete(unsubscribe); err != nil {
		return nil, err
	}
	getLogger().Debug("new subscriber consumer created")

	go consumer.Listen(stopSignal)
	return client, nil
}

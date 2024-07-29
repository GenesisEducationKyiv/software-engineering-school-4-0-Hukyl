package rate

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

func getLogger() *slog.Logger {
	if logger == nil {
		logger = slog.Default().With(slog.Any("src", "rateClient"))
	}
	return logger
}

var rateTimeout = 5 * time.Second

func getTotalReceivedMessagesCounter(eventName string) *metrics.Counter {
	// ? is it better to rework message delivering
	// ? logic to decrease the number of received messages?
	return metrics.GetOrCreateCounter(fmt.Sprintf(
		`broker_received_messages_total{consumer="rate_consumer", event="%s"}`, eventName,
	))
}

const eventType = "RateFetched"

type rateFetchedEvent struct {
	broker.Event
	Data struct {
		From string    `json:"from"`
		To   string    `json:"to"`
		Rate float32   `json:"body"`
		Time time.Time `json:"time"`
	} `json:"data"`
}

type Handler func(ctx context.Context, from, to string, rate float32, time time.Time) error

type Client struct {
	consumer   *transport.Consumer
	stopSignal chan struct{}
}

func (c *Client) Subscribe(f Handler) error {
	slog.Debug("subscribing to rate events", slog.Any("handler", f))
	c.consumer.Subscribe(func(b []byte) error {
		getTotalReceivedMessagesCounter(eventType).Inc()
		getLogger().Debug("received message", slog.Any("handler", f))
		ctx, cancel := context.WithTimeout(context.Background(), rateTimeout)
		defer cancel()
		event, err := c.unmarshal(b)
		if err != nil {
			return err
		}
		if event.Event.Type != eventType {
			return nil
		}
		slog.Info(
			"delivering message",
			slog.Any("listener", f),
			slog.Any("eventName", event.Event.Type),
		)
		return f(ctx, event.Data.From, event.Data.To, event.Data.Rate, event.Data.Time)
	})
	return nil
}

func (c *Client) unmarshal(data []byte) (*rateFetchedEvent, error) {
	event := &rateFetchedEvent{}
	if err := json.Unmarshal(data, event); err != nil {
		return nil, err
	}
	return event, nil
}

func (c *Client) Close() error {
	getLogger().Info("closing rate client")
	close(c.stopSignal)
	return c.consumer.Close()
}

func NewClient(config config.Config) (*Client, error) {
	consumer, err := transport.NewConsumer(config)
	if err != nil {
		slog.Error("creating rate consumer", slog.Any("error", err))
		return nil, err
	}
	getLogger().Debug("new rate consumer created")
	stopSignal := make(chan struct{})
	go consumer.Listen(stopSignal)
	return &Client{consumer: consumer, stopSignal: stopSignal}, nil
}
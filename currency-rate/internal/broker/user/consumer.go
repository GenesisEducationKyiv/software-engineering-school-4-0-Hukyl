package user

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker/transport"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker/transport/config"
)

type Consumer struct {
	consumer   *transport.Consumer
	stopSignal chan struct{}
}

type Handler func(ctx context.Context, email string) error

func (c *Consumer) unmarshal(b []byte) (subscribeEvent, error) {
	var event subscribeEvent
	if err := json.Unmarshal(b, &event); err != nil {
		return subscribeEvent{}, fmt.Errorf("unmarshalling event: %w", err)
	}
	return event, nil
}

func (c *Consumer) handleWithEvent(eventName string, f Handler) func([]byte) error {
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
		slog.Info(
			"delivering message",
			slog.Any("listener", f),
			slog.Any("eventName", event.Event.Type),
		)
		return f(ctx, event.Data.Email)
	}
}

func (c *Consumer) ListenSubscribeCompensate(f Handler) error {
	c.consumer.Subscribe(c.handleWithEvent(compensateSubscribedEventType, f))
	return nil
}

func (c *Consumer) ListenUnsubscribeCompensate(f Handler) error {
	c.consumer.Subscribe(c.handleWithEvent(compensateUnsubscribedEventType, f))
	return nil
}

func (c *Consumer) Start() {
	c.consumer.Listen(c.stopSignal)
}

func (c *Consumer) Close() error {
	close(c.stopSignal)
	return c.consumer.Close()
}

func NewConsumer(config config.Config) (*Consumer, error) {
	consumer, err := transport.NewConsumer(config)
	if err != nil {
		return nil, fmt.Errorf("creating consumer: %w", err)
	}
	stopSignal := make(chan struct{})
	return &Consumer{consumer: consumer, stopSignal: stopSignal}, nil
}

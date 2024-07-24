package subscriber

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker/transport"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker/transport/config"
)

type CompensateClient struct {
	Client
	producer *transport.Producer
}

const (
	compensateSubscribedEventType   = "CompensateSubscribe"
	compensateUnsubscribedEventType = "CompensateUnsubscribe"
)

func (c *CompensateClient) handleWithEvent(
	eventName, compensateEventName string, f Handler,
) func([]byte) error {
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
		if err = f(ctx, event.Data.Email); err != nil {
			compensateEvent := SubscribeEvent{
				Event: event.Event,
				Data:  event.Data,
			}
			compensateEvent.Type = compensateEventName
			c.compensate(compensateEvent)
		}
		return nil
	}
}

func (c *CompensateClient) subscribeCreate(f Handler) error {
	c.consumer.Subscribe(c.handleWithEvent(
		subscribedEventType,
		compensateSubscribedEventType,
		f,
	))
	return nil
}

func (c *CompensateClient) subscribeDelete(f Handler) error {
	c.consumer.Subscribe(c.handleWithEvent(
		unsubscribedEventType,
		compensateUnsubscribedEventType,
		f,
	))
	return nil
}

func (c *CompensateClient) Close() error {
	defer func() {
		if err := c.Client.Close(); err != nil {
			slog.Error("closing client", slog.Any("error", err))
		}
	}()
	return c.producer.Close()
}

func (c *CompensateClient) marshal(event SubscribeEvent) ([]byte, error) {
	return json.Marshal(event)
}

func (c *CompensateClient) compensate(event SubscribeEvent) {
	getLogger().Info("compensating", slog.Any("eventName", event.Type))
	eventBytes, err := c.marshal(event)
	if err != nil {
		slog.Error("compensating", slog.Any("error", err))
		return
	}
	if err := c.producer.Produce(context.Background(), eventBytes); err != nil {
		slog.Error("compensating", slog.Any("error", err))
	}
}

func NewCompensateClient(
	config config.Config, compensateConfig config.Config,
	subscribe Handler, unsubscribe Handler,
) (*CompensateClient, error) {
	consumer, err := transport.NewConsumer(config)
	if err != nil {
		return nil, err
	}
	stopSignal := make(chan struct{})

	client := &Client{consumer: consumer, stopSignal: stopSignal}
	producer, err := transport.NewProducer(compensateConfig)
	if err != nil {
		getLogger().Error("creating producer", slog.Any("error", err))
		return nil, err
	}
	compensateClient := &CompensateClient{Client: *client, producer: producer}

	if err = compensateClient.subscribeCreate(subscribe); err != nil {
		return nil, err
	}
	if err = compensateClient.subscribeDelete(unsubscribe); err != nil {
		return nil, err
	}
	getLogger().Debug("new subscriber consumer with compensation created")

	go consumer.Listen(stopSignal)
	return compensateClient, nil
}

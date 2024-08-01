package user

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker/transport"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker/transport/config"
	"github.com/VictoriaMetrics/metrics"
)

var lastEventID int

var producerLogger *slog.Logger

func getProducerLogger() *slog.Logger {
	if producerLogger == nil {
		producerLogger = slog.Default().With(slog.Any("src", "userProducer"))
	}
	return producerLogger
}

var (
	totalSentUserMessagesCounter = metrics.NewCounter(
		`broker_sent_messages_total{producer="user_producer"}`,
	)
	messageSizeHistogram = metrics.NewHistogram(
		`broker_message_size_bytes{producer="user_producer"}`,
	)
)

type Producer struct {
	producer *transport.Producer
}

func (p *Producer) createEvent(eventName string, data subscriberData) subscribeEvent {
	lastEventID++
	return subscribeEvent{
		Event: broker.Event{
			ID:        strconv.Itoa(lastEventID),
			Type:      eventName,
			Timestamp: time.Now().Format(time.RFC3339),
		},
		Data: data,
	}
}

func (p *Producer) marshal(data subscribeEvent) ([]byte, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshalling command: %w", err)
	}
	return bytes, nil
}

func (p *Producer) sendEvent(
	ctx context.Context,
	eventName string,
	data subscriberData,
) error {
	event := p.createEvent(eventName, data)
	getProducerLogger().Info(
		"producing user event",
		slog.Any("event", eventName),
	)
	msgBytes, err := p.marshal(event)
	if err != nil {
		return err
	}
	if err := p.producer.Produce(ctx, msgBytes); err != nil {
		return fmt.Errorf("producing event: %w", err)
	}
	totalSentUserMessagesCounter.Inc()
	messageSizeHistogram.Update(float64(len(msgBytes)))
	return nil
}

func (p *Producer) SendSubscribe(
	ctx context.Context,
	email string,
) error {
	err := p.sendEvent(ctx, subscribedEventType, subscriberData{
		Email: email,
	})
	if err != nil {
		return fmt.Errorf("sending subscribe event: %w", err)
	}
	return nil
}

func (p *Producer) SendUnsubscribe(
	ctx context.Context,
	email string,
) error {
	err := p.sendEvent(ctx, unsubscribedEventType, subscriberData{
		Email: email,
	})
	if err != nil {
		return fmt.Errorf("sending unsubscribe event: %w", err)
	}
	return nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}

func NewProducer(config config.Config) (*Producer, error) {
	producer, err := transport.NewProducer(config)
	if err != nil {
		return nil, fmt.Errorf("creating producer: %w", err)
	}
	return &Producer{producer}, nil
}

package rate

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

var logger *slog.Logger

var (
	totalSentRateMessagesCounter = metrics.NewCounter(
		`broker_sent_messages_total{producer="rate_producer"}`,
	)
	messageSizeHistogram = metrics.NewHistogram(
		`broker_message_size_bytes{producer="rate_producer"}`,
	)
)

func getLogger() *slog.Logger {
	if logger == nil {
		logger = slog.Default().With(slog.Any("src", "rateProducer"))
	}
	return logger
}

type Producer struct {
	producer *transport.Producer
}

func (m *Producer) createEvent(data rateData) rateFetchedEvent {
	lastEventID++
	return rateFetchedEvent{
		Event: broker.Event{
			ID:        strconv.Itoa(lastEventID),
			Type:      eventType,
			Timestamp: time.Now().Format(time.RFC3339),
		},
		Data: data,
	}
}

func (m *Producer) marshal(data rateFetchedEvent) ([]byte, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshalling command: %w", err)
	}
	return bytes, nil
}

func (m *Producer) SendRate(
	ctx context.Context,
	from string, to string, rate float32,
) error {
	data := m.createEvent(rateData{
		From: from,
		To:   to,
		Rate: rate,
		Time: time.Now(),
	})
	getLogger().Info("producing rate message", slog.Any("rate", rate))
	msgBytes, err := m.marshal(data)
	if err != nil {
		return err
	}
	if err := m.producer.Produce(ctx, msgBytes); err != nil {
		return fmt.Errorf("producing rate message: %w", err)
	}
	totalSentRateMessagesCounter.Inc()
	messageSizeHistogram.Update(float64(len(msgBytes)))
	getLogger().Debug("rate message produced")
	return nil
}

func (m *Producer) Close() error {
	return m.producer.Close()
}

func NewProducer(config config.Config) (*Producer, error) {
	producer, err := transport.NewProducer(config)
	if err != nil {
		return nil, fmt.Errorf("creating producer: %w", err)
	}
	return &Producer{producer}, nil
}
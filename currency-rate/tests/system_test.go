//go:build system

package tests_test

import (
	"context"
	"testing"

	rateProducer "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/broker/rate"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker/transport"
	transportCfg "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker/transport/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NOTE: This test requires a running RabbitMQ instance.
func TestRateProducer_ValidMessage(t *testing.T) {
	// Arrange
	rp, err := rateProducer.NewProducer(transportCfg.NewFromEnv())
	require.NoError(t, err)
	consumer, err := transport.NewConsumer(transportCfg.NewFromEnv())
	require.NoError(t, err)
	done := make(chan struct{})
	messagesReceived := 0

	listener := func(_ []byte) error {
		defer func() { done <- struct{}{} }()
		messagesReceived++
		return nil
	}

	// Act
	consumer.Subscribe(listener)
	go consumer.Listen(done)
	err = rp.SendRate(context.Background(), "USD", "UAH", 27.5)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, 1, messagesReceived)
}

//go:build system

package tests_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"

	rateProducer "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/broker/rate"
	userBroker "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/broker/user"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/handler"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/models"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker/transport"
	transportCfg "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker/transport/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/database"
	databaseCfg "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/database/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/settings"
	"github.com/gin-gonic/gin"
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
	time.Sleep(1 * time.Second)
	// Assert
	assert.Equal(t, 1, messagesReceived)
}

func TestUserSubscriptionSaga_Compensate(t *testing.T) {
	// Arrange
	// Create user repo saga
	db, err := database.New(databaseCfg.NewFromEnv())
	require.NoError(t, err)
	err = db.Migrate(&models.User{})
	require.NoError(t, err)
	userRepo := models.NewUserRepository(db)
	brokerURI := transportCfg.NewFromEnv().BrokerURI
	slog.Debug(brokerURI)
	userProducer, err := userBroker.NewProducer(transportCfg.Config{
		BrokerURI: brokerURI,
		QueueName: "user",
	})
	require.NoError(t, err)
	userConsumer, err := userBroker.NewConsumer(transportCfg.Config{
		BrokerURI: brokerURI,
		QueueName: "user_compensate",
	})
	require.NoError(t, err)
	defer userConsumer.Close()
	slog.Debug("initialized userConsumer and userProducer")
	// Decorate userRepo with userProducer
	urs := handler.NewUserRepositorySaga(
		userRepo,
		userProducer,
		userConsumer,
	)
	slog.Debug("initialized saga")
	email := "example@gmail.com"
	// Create compensator
	consumer, err := transport.NewConsumer(transportCfg.Config{
		BrokerURI: brokerURI,
		QueueName: "user",
	})
	require.NoError(t, err)
	producer, err := transport.NewProducer(transportCfg.Config{
		BrokerURI: brokerURI,
		QueueName: "user_compensate",
	})
	slog.Debug("initialized compensator's producer and consumer")
	require.NoError(t, err)
	done := make(chan struct{})
	listener := func(msg []byte) error {
		var eventData broker.Event
		err := json.Unmarshal(msg, &eventData)
		require.NoError(t, err)
		eventData.Type = "Compensate" + eventData.Type // FIXME: hardcoded compensate event type
		newMsgBytes, err := json.Marshal(eventData)
		require.NoError(t, err)
		return producer.Produce(context.Background(), newMsgBytes)
	}
	// Act
	consumer.Subscribe(listener)
	go consumer.Listen(done)
	defer func() { done <- struct{}{} }()
	slog.Debug("subscribed compensator")
	slog.Debug("creating user...")
	err = urs.Create(&models.User{Email: email})
	require.NoError(t, err)
	slog.Debug("sleeping...")
	time.Sleep(5 * time.Second)
	// Assert
	slog.Debug("fetching users...")
	users, err := userRepo.FindAll()
	require.NoError(t, err)
	assert.Empty(t, users)
}

func TestMain(m *testing.M) {
	// Run tests
	gin.SetMode(gin.ReleaseMode)
	_ = settings.InitSettings()
	slog.SetDefault(
		slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		),
	)
	m.Run()
}

package handlers_test

import (
	"context"
	"testing"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/handlers"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type subscriberRepositoryMock struct {
	mock.Mock
}

func (m *subscriberRepositoryMock) Create(subscriber *models.Subscriber) error {
	args := m.Called(subscriber)
	return args.Error(0)
}

func (m *subscriberRepositoryMock) Delete(subscriber *models.Subscriber) error {
	args := m.Called(subscriber)
	return args.Error(0)
}

func TestSubscriberEvents_Subscribe(t *testing.T) {
	// Arrange
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	repo := new(subscriberRepositoryMock)
	repo.On("Create", mock.Anything).Return(nil)
	subscriberEvents := handlers.NewSubscriberEvents(repo)
	// Act
	err := subscriberEvents.Subscribe(ctx, "example@gmail.com")
	require.NoError(t, err)
	// Assert
	repo.AssertCalled(t, "Create", mock.MatchedBy(func(subscriber *models.Subscriber) bool {
		return subscriber.Email == "example@gmail.com"
	}))
}

func TestSubscriberEvents_Unsubscribe(t *testing.T) {
	// Arrange
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	repo := new(subscriberRepositoryMock)
	repo.On("Delete", mock.Anything).Return(nil)
	subscriberEvents := handlers.NewSubscriberEvents(repo)
	// Act
	err := subscriberEvents.Unsubscribe(ctx, "example@gmail.com")
	// Assert
	require.NoError(t, err)
	repo.AssertCalled(t, "Delete", mock.MatchedBy(func(subscriber *models.Subscriber) bool {
		return subscriber.Email == "example@gmail.com"
	}))
}

func TestSubscriberEvents_CancelledContext(t *testing.T) {
	// Arrange
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	repo := new(subscriberRepositoryMock)
	repo.On("Create", mock.Anything).Return(nil)
	subscriberEvents := handlers.NewSubscriberEvents(repo)
	// Act
	err := subscriberEvents.Subscribe(ctx, "test@gmail.com")
	// Assert
	require.Error(t, err)
	repo.AssertNotCalled(t, "Create")
}

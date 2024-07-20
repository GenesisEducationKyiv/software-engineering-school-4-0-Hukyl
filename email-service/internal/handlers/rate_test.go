package handlers_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/handlers"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type rateRepositoryMock struct {
	mock.Mock
}

func (m *rateRepositoryMock) Create(rate *models.Rate) error {
	args := m.Called(rate)
	return args.Error(0)
}

func TestRateEvents_SaveRate(t *testing.T) {
	// Arrange
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	repo := new(rateRepositoryMock)
	repo.On("Create", mock.Anything).Return(nil)
	rateEvents := handlers.NewRateEvents(repo)
	// Act
	err := rateEvents.SaveRate(ctx, "USD", "UAH", 27.5, time.Now())
	// Assert
	require.NoError(t, err)
	repo.AssertCalled(t, "Create", mock.MatchedBy(func(rate *models.Rate) bool {
		validFrom := rate.CurrencyFrom == "USD"
		validTo := rate.CurrencyTo == "UAH"
		validRate := math.Abs(float64(rate.Rate-27.5)) < 0.0001
		return validFrom && validTo && validRate
	}))
}

func TestRateEvents_CancelledContext(t *testing.T) {
	// Arrange
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	repo := new(rateRepositoryMock)
	rateEvents := handlers.NewRateEvents(repo)
	// Act
	err := rateEvents.SaveRate(ctx, "USD", "UAH", 27.5, time.Now())
	// Assert
	require.Error(t, err)
	repo.AssertNotCalled(t, "Create")
}

package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/rate"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/server/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type (
	mockRateFetcher struct {
		mock.Mock
	}
)

func (m *mockRateFetcher) FetchRate(ctx context.Context, from, to string) (rate.Rate, error) {
	args := m.Called(ctx, from, to)
	return args.Get(0).(rate.Rate), args.Error(1)
}

func TestFetchRate(t *testing.T) {
	// Arrange
	ccFrom := "USD"
	ccTo := "UAH"
	expected := rate.Rate{
		CurrencyFrom: ccFrom,
		CurrencyTo:   ccTo,
		Rate:         27.5,
	}
	firstFetcher := new(mockRateFetcher)
	firstFetcher.On("FetchRate", mock.Anything, ccFrom, ccTo).Return(
		rate.Rate{}, errors.New("failed to fetch rate"),
	)

	secondFetcher := new(mockRateFetcher)
	secondFetcher.On("FetchRate", mock.Anything, ccFrom, ccTo).Return(expected, nil)

	s := service.NewRateService(firstFetcher, secondFetcher)

	// Act
	result, err := s.FetchRate(context.Background(), ccFrom, ccTo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expected, result)
	firstFetcher.AssertExpectations(t)
	secondFetcher.AssertExpectations(t)
}

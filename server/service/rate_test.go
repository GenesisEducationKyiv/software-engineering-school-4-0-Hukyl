package service_test

import (
	"testing"

	"github.com/Hukyl/genesis-kma-school-entry/models"
	"github.com/Hukyl/genesis-kma-school-entry/rate"
	"github.com/Hukyl/genesis-kma-school-entry/server/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type (
	mockRateFetcher struct {
		mock.Mock
	}

	mockRateRepository struct {
		mock.Mock
	}
)

func (m *mockRateFetcher) FetchRate(from, to string) (rate.Rate, error) {
	args := m.Called(from, to)
	return args.Get(0).(rate.Rate), args.Error(1)
}

func (m *mockRateRepository) Create(rate *models.Rate) error {
	args := m.Called(rate)
	return args.Error(0)
}

func TestFetchRate(t *testing.T) {
	// Arrange
	expected := &models.Rate{
		CurrencyFrom: "USD",
		CurrencyTo:   "UAH",
		Rate:         27.5,
	}
	mockFetcher := new(mockRateFetcher)
	mockFetcher.On("FetchRate", "USD", "UAH").Return(rate.Rate{
		CurrencyFrom: expected.CurrencyFrom,
		CurrencyTo:   expected.CurrencyTo,
		Rate:         expected.Rate,
	}, nil)

	mockRepo := new(mockRateRepository)
	mockRepo.On("Create", expected).Return(nil)

	s := service.NewRateService(mockRepo, mockFetcher)

	// Act
	result, err := s.FetchRate("USD", "UAH")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	mockFetcher.AssertExpectations(t)
}

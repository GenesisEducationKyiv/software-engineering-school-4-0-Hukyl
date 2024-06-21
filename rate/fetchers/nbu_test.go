package fetchers_test

import (
	"testing"

	"github.com/Hukyl/genesis-kma-school-entry/rate/fetchers"
	"github.com/stretchr/testify/assert"
)

func TestNBUUnsupportedCurrency(t *testing.T) {
	nbu := fetchers.NewNBURateFetcher()
	_, err := nbu.FetchRate("-", "UAH")
	assert.Error(t, err)
}

func TestNBUFetchRate(t *testing.T) {
	nbu := fetchers.NewNBURateFetcher()
	rate, err := nbu.FetchRate("USD", "UAH")
	assert.NoError(t, err)
	assert.Greater(t, rate.Rate, float32(0))
}

func TestNBUOnlyUAH(t *testing.T) {
	nbu := fetchers.NewNBURateFetcher()
	_, err := nbu.FetchRate("USD", "EUR")
	assert.Error(t, err)
}

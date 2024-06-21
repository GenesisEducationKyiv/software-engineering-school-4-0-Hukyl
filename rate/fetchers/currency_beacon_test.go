package fetchers_test

import (
	"testing"

	"github.com/Hukyl/genesis-kma-school-entry/rate/fetchers"
	"github.com/stretchr/testify/assert"
)

func TestCurrencyBeaconUnsupportedCurrency(t *testing.T) {
	b := fetchers.NewCurrencyBeaconFetcher("")
	_, err := b.FetchRate("-", "UAH")
	assert.Error(t, err)
}

func TestCurrencyBeaconNoAPIKey(t *testing.T) {
	b := fetchers.NewCurrencyBeaconFetcher("")
	_, err := b.FetchRate("USD", "EUR")
	assert.Error(t, err)
}

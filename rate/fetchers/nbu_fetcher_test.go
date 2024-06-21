package fetchers_test

import (
	"testing"

	"github.com/Hukyl/genesis-kma-school-entry/rate/fetchers"
)

func TestNBUUnsupportedCurrency(t *testing.T) {
	nbu := fetchers.NewNBURateFetcher()
	_, err := nbu.FetchRate("-", "UAH")
	if err == nil {
		t.Error("expected an error")
	}
}

func TestNBUFetchRate(t *testing.T) {
	nbu := fetchers.NewNBURateFetcher()
	rate, err := nbu.FetchRate("USD", "UAH")
	if err != nil {
		t.Fatal(err)
	}
	if rate.Rate <= 0 {
		t.Error("expected a positive rate")
	}
}

func TestNBUOnlyUAH(t *testing.T) {
	nbu := fetchers.NewNBURateFetcher()
	_, err := nbu.FetchRate("USD", "EUR")
	if err == nil {
		t.Error("expected an error")
	}
}

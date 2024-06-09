package rate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"time"
)

// Fetcher is an interface that defines the methods for fetching exchange rates.
//
// SupportedCurrencies returns a list of currency codes that the Fetcher supports.
//
// FetchRate fetches the exchange rate (either from local storage, by accessing API
// or in any other way) between two currencies.
type Fetcher interface {
	SupportedCurrencies() []string
	FetchRate(ccFrom, ccTo string) (Rate, error)
}

// nbuRateFetcher is a RateFetcher implementation that fetches rates from
// the National Bank of Ukraine
// API docs: https://bank.gov.ua/ua/open-data/api-dev
// NOTE: CurrencyTo can only be "UAH", as the NBU API only supports fetching rates for UAH
//
// Example usage:
//
//	fetcher := NewNBURateFetcher()
//	rate, err := fetcher.FetchRate("USD", "UAH")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(rate)
type nbuRateFetcher struct{}

func (n nbuRateFetcher) SupportedCurrencies() []string {
	return []string{"UAH", "USD"}
}

func (n nbuRateFetcher) formatURL(cc string, date time.Time) string {
	currentDate := fmt.Sprintf("%d%02d%02d", date.Year(), date.Month(), date.Day())
	baseURL := "https://bank.gov.ua/NBUStatService/v1/statdirectory/" +
		"exchange?valcode=%s&date=%s&json"
	return fmt.Sprintf(baseURL, cc, currentDate)
}

func (n nbuRateFetcher) FetchRate(ccFrom, ccTo string) (Rate, error) {
	if ccTo != "UAH" {
		return Rate{}, fmt.Errorf("invalid currency from: %s", ccFrom)
	}
	rate := Rate{
		CurrencyFrom: ccFrom,
		CurrencyTo:   ccTo,
		Time:         time.Now(),
	}
	if !slices.Contains(n.SupportedCurrencies(), ccFrom) {
		return rate, fmt.Errorf("unsupported currency: %s", ccFrom)
	}
	resp, err := http.Get(n.formatURL(ccFrom, time.Now()))
	if err != nil {
		return rate, err
	}
	defer resp.Body.Close()
	var data []struct {
		Rate float32 `json:"rate"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return rate, err
	}
	if len(data) == 0 {
		return rate, fmt.Errorf("no rate data found")
	}
	rate.Rate = data[0].Rate
	return rate, nil
}

func NewNBURateFetcher() Fetcher {
	return nbuRateFetcher{}
}

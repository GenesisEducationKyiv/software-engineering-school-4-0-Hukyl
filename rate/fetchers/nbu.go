package fetchers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/Hukyl/genesis-kma-school-entry/rate"
)

// NBURateFetcher is a RateFetcher implementation that fetches rates from
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
type NBURateFetcher struct {
	next RateFetcher
}

func (n *NBURateFetcher) SupportedCurrencies() []string {
	return []string{"UAH", "USD"}
}

func (n *NBURateFetcher) formatURL(cc string, date time.Time) string {
	currentDate := fmt.Sprintf("%d%02d%02d", date.Year(), date.Month(), date.Day())
	baseURL := "https://bank.gov.ua/NBUStatService/v1/statdirectory/" +
		"exchange?valcode=%s&date=%s&json"
	return fmt.Sprintf(baseURL, cc, currentDate)
}

func (n *NBURateFetcher) fetchRate(ccFrom, ccTo string) (rate.Rate, error) {
	if ccTo != "UAH" {
		return rate.Rate{}, fmt.Errorf("invalid currency from: %s", ccFrom)
	}
	result := rate.Rate{
		CurrencyFrom: ccFrom,
		CurrencyTo:   ccTo,
		Time:         time.Now(),
	}
	if !slices.Contains(n.SupportedCurrencies(), ccFrom) {
		return result, fmt.Errorf("unsupported currency: %s", ccFrom)
	}
	resp, err := http.Get(n.formatURL(ccFrom, time.Now()))
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	var data []struct {
		Rate float32 `json:"rate"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return result, err
	}
	if len(data) == 0 {
		return result, fmt.Errorf("no rate data found")
	}
	result.Rate = data[0].Rate
	return result, nil
}

func (n *NBURateFetcher) FetchRate(ccFrom, ccTo string) (rate.Rate, error) {
	result, err := n.fetchRate(ccFrom, ccTo)
	slog.Info(
		"Fetched rate",
		slog.String("fetcher", fmt.Sprint(n)), slog.Any("rate", result), slog.Any("error", err),
	)
	if err == nil {
		return result, nil
	}
	if n.next != nil {
		return n.next.FetchRate(ccFrom, ccTo)
	}
	return rate.Rate{}, err
}

func (n *NBURateFetcher) SetNext(next RateFetcher) {
	n.next = next
}

func (n *NBURateFetcher) String() string {
	return "NBURateFetcher{}"
}

func NewNBURateFetcher() *NBURateFetcher {
	return &NBURateFetcher{}
}

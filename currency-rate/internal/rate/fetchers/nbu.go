package fetchers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/rate"
)

// NBURateFetcher is a RateFetcher implementation that fetches rates from
// the National Bank of Ukraine
// API docs: https://bank.gov.ua/ua/open-data/api-dev
// NOTE: CurrencyTo can only be "UAH", as the NBU API only supports fetching rates for UAH
//
// Example usage:
//
//	fetcher := NewNBURateFetcher()
//	rate, err := fetcher.FetchRate(context.Background(), "USD", "UAH")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(rate)
type NBURateFetcher struct{}

const uahCC = "UAH"

func (n *NBURateFetcher) SupportedCurrencies(_ context.Context) []string {
	return []string{uahCC, "USD"}
}

func (n *NBURateFetcher) formatURL(cc string, date time.Time) string {
	currentDate := fmt.Sprintf("%d%02d%02d", date.Year(), date.Month(), date.Day())
	const baseURL = "https://bank.gov.ua/NBUStatService/v1/statdirectory/" +
		"exchange?valcode=%s&date=%s&json"
	return fmt.Sprintf(baseURL, cc, currentDate)
}

func (n *NBURateFetcher) fetchRate(ctx context.Context, ccFrom, ccTo string) (rate.Rate, error) {
	if ccTo != uahCC {
		return rate.Rate{}, fmt.Errorf("invalid currency from: %s", ccFrom)
	}
	result := rate.Rate{
		CurrencyFrom: ccFrom,
		CurrencyTo:   ccTo,
		Time:         time.Now(),
	}
	if !slices.Contains(n.SupportedCurrencies(ctx), ccFrom) {
		return result, fmt.Errorf("unsupported currency: %s", ccFrom)
	}
	formattedURL := n.formatURL(ccFrom, time.Now())
	req, err := http.NewRequest(http.MethodGet, formattedURL, nil)
	if err != nil {
		return result, err
	}
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
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
		return result, errors.New("no rate data found")
	}
	result.Rate = data[0].Rate
	return result, nil
}

func (n *NBURateFetcher) FetchRate(ctx context.Context, ccFrom, ccTo string) (rate.Rate, error) {
	result, err := n.fetchRate(ctx, ccFrom, ccTo)
	slog.Info(
		"fetched rate",
		slog.String("fetcher", fmt.Sprint(n)), slog.Any("rate", result), slog.Any("error", err),
	)
	if err == nil {
		return result, nil
	}
	return rate.Rate{}, err
}

func (n *NBURateFetcher) String() string {
	return "NBURateFetcher{}"
}

func NewNBURateFetcher() *NBURateFetcher {
	return &NBURateFetcher{}
}

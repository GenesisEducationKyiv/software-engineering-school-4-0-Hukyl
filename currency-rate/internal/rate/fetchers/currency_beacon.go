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
	"github.com/VictoriaMetrics/metrics"
	"github.com/ericchiang/css"
	"golang.org/x/net/html"
)

const (
	cssSelector            = "tbody tr > td:first-child > a"
	supportedCurrenciesURL = "https://currencybeacon.com/supported-currencies"
	baseURL                = "https://api.currencybeacon.com/v1/latest?" +
		"api_key=%s&base=%s&symbols=%s"
)

var (
	beaconConsecutiveErrorsMetric = metrics.GetOrCreateCounter(
		`rate_fetcher_consecutive_errors_total{fetcher="currency_beacon_fetcher"}`,
	)
	beaconResponseTimeMetric = metrics.GetOrCreateHistogram(
		`rate_fetcher_response_duration_seconds{fetcher="currency_beacon_fetcher"}`,
	)
)

type endpointResponse struct {
	Rates map[string]float32 `json:"rates"`
}

type CurrencyBeaconFetcher struct {
	APIKey              string
	supportedCurrencies []string
}

func (c *CurrencyBeaconFetcher) SupportedCurrencies(ctx context.Context) []string {
	if c.supportedCurrencies != nil {
		return c.supportedCurrencies
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, supportedCurrenciesURL, nil)
	if err != nil {
		getLogger().Error(
			"creating request",
			slog.String("fetcher", fmt.Sprint(c)), slog.Any("error", err),
		)
		return nil
	}
	startTime := time.Now()
	response, err := http.DefaultClient.Do(req)
	getLogger().Info(
		"fetching supported currencies",
		slog.String("fetcher", fmt.Sprint(c)), slog.Any("error", err),
	)
	if err != nil {
		return nil
	}
	defer response.Body.Close()
	beaconResponseTimeMetric.UpdateDuration(startTime)
	sel, _ := css.Parse(cssSelector)
	node, err := html.Parse(response.Body)
	if err != nil {
		getLogger().Error(
			"parsing html",
			slog.String("fetcher", fmt.Sprint(c)), slog.Any("error", err),
		)
		return nil
	}
	currencies := make([]string, 0, 100)
	for _, n := range sel.Select(node) {
		currencies = append(currencies, n.FirstChild.Data)
	}
	c.supportedCurrencies = currencies
	return currencies
}

func (c *CurrencyBeaconFetcher) fetchRate(
	ctx context.Context, ccFrom, ccTo string,
) (rate.Rate, error) {
	formattedURL := fmt.Sprintf(baseURL, c.APIKey, ccFrom, ccTo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, formattedURL, nil)
	if err != nil {
		return rate.Rate{}, err
	}
	startTime := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return rate.Rate{}, err
	}
	defer resp.Body.Close()
	beaconResponseTimeMetric.UpdateDuration(startTime)
	getLogger().Debug(
		"fetched rate",
		slog.String("url", formattedURL),
		slog.Any("status", resp.Status),
	)
	if resp.StatusCode != http.StatusOK {
		return rate.Rate{}, fmt.Errorf("fetching url: %s", resp.Status)
	}
	var data endpointResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return rate.Rate{}, err
	}
	value, ok := data.Rates[ccTo]
	if !ok {
		return rate.Rate{}, errors.New("rate not found")
	}
	rate := rate.Rate{
		CurrencyFrom: ccFrom,
		CurrencyTo:   ccTo,
		Rate:         value,
		Time:         time.Now(),
	}
	getLogger().Debug("rate fetched", slog.Any("rate", rate))
	return rate, nil
}

func (c *CurrencyBeaconFetcher) FetchRate(
	ctx context.Context, ccFrom, ccTo string,
) (rate.Rate, error) {
	supportedCurrencies := c.SupportedCurrencies(ctx)
	if supportedCurrencies == nil {
		err := errors.New("failed to fetch supported currencies")
		getLogger().Info(
			"fetching rate",
			slog.String("fetcher", fmt.Sprint(c)),
			slog.Any("error", err),
		)
		return rate.Rate{}, err
	}
	if !slices.Contains(supportedCurrencies, ccFrom) {
		return rate.Rate{}, fmt.Errorf("unsupported currency: %s", ccFrom)
	}
	if !slices.Contains(supportedCurrencies, ccTo) {
		return rate.Rate{}, fmt.Errorf("unsupported currency: %s", ccTo)
	}
	result, err := c.fetchRate(ctx, ccFrom, ccTo)
	getLogger().Info(
		"fetched rate",
		slog.String("fetcher", fmt.Sprint(c)),
		slog.Any("rate", result),
		slog.Any("error", err),
	)
	if err == nil {
		beaconConsecutiveErrorsMetric.Set(0)
		return result, nil
	}
	beaconConsecutiveErrorsMetric.Inc()
	return rate.Rate{}, err
}

func (c *CurrencyBeaconFetcher) String() string {
	return "CurrencyBeaconFetcher{}"
}

func NewCurrencyBeaconFetcher(apiKey string) *CurrencyBeaconFetcher {
	return &CurrencyBeaconFetcher{
		APIKey: apiKey,
	}
}

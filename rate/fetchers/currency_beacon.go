package fetchers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/Hukyl/genesis-kma-school-entry/rate"
	"github.com/ericchiang/css"
	"golang.org/x/net/html"
)

const (
	cssSelector            string = "tbody tr > td:first-child > a"
	supportedCurrenciesURL string = "https://currencybeacon.com/supported-currencies"
	baseURL                string = "https://api.currencybeacon.com/v1/latest?api_key=%s&base=%s&symbols=%s" // nolint: lll

)

type CurrencyBeaconFetcher struct {
	APIKey              string
	next                RateFetcher
	supportedCurrencies []string
}

func (c *CurrencyBeaconFetcher) SupportedCurrencies() []string {
	if c.supportedCurrencies != nil {
		return c.supportedCurrencies
	}
	request, err := http.Get(supportedCurrenciesURL)
	slog.Info(
		"fetching supported currencies",
		slog.String("fetcher", fmt.Sprint(c)), slog.Any("error", err),
	)
	if err != nil {
		return nil
	}
	defer request.Body.Close()
	sel, _ := css.Parse(cssSelector)
	node, err := html.Parse(request.Body)
	if err != nil {
		slog.Error(
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

func (c *CurrencyBeaconFetcher) fetchRate(ccFrom, ccTo string) (rate.Rate, error) {
	formattedURL := fmt.Sprintf(baseURL, c.APIKey, ccFrom, ccTo)
	req, err := http.NewRequest(http.MethodGet, formattedURL, nil)
	if err != nil {
		return rate.Rate{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return rate.Rate{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return rate.Rate{}, fmt.Errorf("fetching url: %s", resp.Status)
	}
	var data struct {
		Rates map[string]float32 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return rate.Rate{}, err
	}
	value, ok := data.Rates[ccTo]
	if !ok {
		return rate.Rate{}, fmt.Errorf("rate not found")
	}
	return rate.Rate{
		CurrencyFrom: ccFrom,
		CurrencyTo:   ccTo,
		Rate:         value,
		Time:         time.Now(),
	}, nil
}

func (c *CurrencyBeaconFetcher) FetchRate(ccFrom, ccTo string) (rate.Rate, error) {
	supportedCurrencies := c.SupportedCurrencies()
	if supportedCurrencies == nil {
		err := fmt.Errorf("failed to fetch supported currencies")
		slog.Info(
			"fetching rate",
			slog.String("fetcher", fmt.Sprint(c)),
			slog.Any("error", err),
		)
		return rate.Rate{}, err
	}
	if !slices.Contains(c.SupportedCurrencies(), ccFrom) {
		return rate.Rate{}, fmt.Errorf("unsupported currency: %s", ccFrom)
	}
	if !slices.Contains(c.SupportedCurrencies(), ccTo) {
		return rate.Rate{}, fmt.Errorf("unsupported currency: %s", ccTo)
	}
	result, err := c.fetchRate(ccFrom, ccTo)
	slog.Info(
		"fetched rate",
		slog.String("fetcher", fmt.Sprint(c)),
		slog.Any("rate", result),
		slog.Any("error", err),
	)
	if err == nil {
		return result, nil
	}
	if c.next != nil {
		return c.next.FetchRate(ccFrom, ccTo)
	}
	return rate.Rate{}, err
}

func (c *CurrencyBeaconFetcher) SetNext(r RateFetcher) {
	c.next = r
}

func (c *CurrencyBeaconFetcher) String() string {
	return "CurrencyBeaconFetcher{}"
}

func NewCurrencyBeaconFetcher(apiKey string) *CurrencyBeaconFetcher {
	return &CurrencyBeaconFetcher{
		APIKey: apiKey,
	}
}

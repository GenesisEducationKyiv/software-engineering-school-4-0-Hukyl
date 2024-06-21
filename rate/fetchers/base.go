package fetchers

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/Hukyl/genesis-kma-school-entry/rate"
)

var ErrRateChainEnd = errors.New("rate chain end")

// RateFetcher is an interface that defines the general behavior of a rate fetcher.
// This fetcher interfaces presumes the use of Chain of Responsibility pattern.
type RateFetcher interface {
	FetchRate(ccFrom, ccTo string) (rate.Rate, error)
	SetNext(next RateFetcher)
}

type BaseFetcher struct {
	next RateFetcher
}

func (b *BaseFetcher) SetNext(r RateFetcher) {
	b.next = r
}

func (b *BaseFetcher) String() string {
	return "BaseFetcher{}"
}

func (b *BaseFetcher) FetchRate(ccFrom, ccTo string) (rate.Rate, error) {
	result := rate.Rate{
		CurrencyFrom: ccFrom,
		CurrencyTo:   ccTo,
	}
	slog.Info(
		"fetching rate", slog.String("fetcher", fmt.Sprint(b)),
		slog.Any("rate", result), slog.Any("error", ErrRateChainEnd),
	)
	return result, ErrRateChainEnd
}

func NewBaseFetcher() *BaseFetcher {
	return &BaseFetcher{}
}

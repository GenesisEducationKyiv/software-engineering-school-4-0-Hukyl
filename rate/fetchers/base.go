package fetchers

import (
	"context"

	"github.com/Hukyl/genesis-kma-school-entry/rate"
)

// RateFetcher is an interface that defines the general behavior of a rate fetcher.
// This fetcher interfaces presumes the use of Chain of Responsibility pattern.
type RateFetcher interface {
	FetchRate(ctx context.Context, ccFrom, ccTo string) (rate.Rate, error)
	SetNext(next RateFetcher)
}

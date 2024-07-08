package rate

import (
	"context"
	"fmt"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/rate"
)

var rateTimeout = 10 * time.Second

type Fetcher interface {
	FetchRate(ctx context.Context, ccFrom, ccTo string) (rate.Rate, error)
}

type Producer interface {
	SendRate(ctx context.Context, from string, to string, rate float32) error
}

type CronJob struct {
	fetcher  Fetcher
	producer Producer
	ccFrom   string
	ccTo     string
}

func (r *CronJob) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), rateTimeout)
	defer cancel()
	rates, err := r.fetcher.FetchRate(ctx, r.ccFrom, r.ccTo)
	if err != nil {
		return fmt.Errorf("fetching rate: %w", err)
	}

	err = r.producer.SendRate(
		ctx,
		rates.CurrencyFrom,
		rates.CurrencyTo,
		rates.Rate,
	)
	if err != nil {
		return fmt.Errorf("sending rate: %w", err)
	}

	return nil
}

func NewCronJob(fetcher Fetcher, producer Producer, ccFrom, ccTo string) *CronJob {
	return &CronJob{
		fetcher:  fetcher,
		producer: producer,
		ccFrom:   ccFrom,
		ccTo:     ccTo,
	}
}

package rate

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/rate"
)

var rateTimeout = 10 * time.Second

var logger *slog.Logger

func getLogger() *slog.Logger {
	if logger == nil {
		logger = slog.Default().With(slog.Any("src", "rateCronJob"))
	}
	return logger
}

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
	getLogger().Debug("rate cron job triggered")
	ctx, cancel := context.WithTimeout(context.Background(), rateTimeout)
	defer cancel()
	rates, err := r.fetcher.FetchRate(ctx, r.ccFrom, r.ccTo)
	if err != nil {
		return fmt.Errorf("fetching rate: %w", err)
	}
	getLogger().Debug("rate fetched", slog.Any("rate", rates))

	err = r.producer.SendRate(
		ctx,
		rates.CurrencyFrom,
		rates.CurrencyTo,
		rates.Rate,
	)
	if err != nil {
		return fmt.Errorf("sending rate: %w", err)
	}
	getLogger().Debug("rate sent", slog.Any("rate", rates))

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

package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/rate"
)

var logger *slog.Logger

func getLogger() *slog.Logger {
	if logger == nil {
		logger = slog.Default().With(slog.Any("src", "service"))
	}
	return logger
}

type RateFetcher interface {
	FetchRate(ctx context.Context, ccFrom, ccTo string) (rate.Rate, error)
}

type RateService struct {
	fetchers []RateFetcher
}

var _ RateFetcher = (*RateService)(nil) // Ensure RateService implements RateFetcher

func (s *RateService) FetchRate(ctx context.Context, from, to string) (rate.Rate, error) {
	for _, f := range s.fetchers {
		getLogger().Debug("fetching rate", slog.Any("fetcher", f))
		r, err := f.FetchRate(ctx, from, to)
		if err != nil {
			getLogger().Warn("fetching rate", slog.Any("fetcher", f), slog.Any("error", err))
			continue
		}
		return r, nil
	}
	getLogger().Error("failed to fetch rate")
	return rate.Rate{}, errors.New("failed to fetch rate")
}

func (s *RateService) SetNext(f ...RateFetcher) {
	getLogger().Debug("adding next fetcher", slog.Any("fetcher", f))
	s.fetchers = append(s.fetchers, f...)
}

func NewRateService(fetcher ...RateFetcher) *RateService {
	return &RateService{
		fetchers: fetcher,
	}
}

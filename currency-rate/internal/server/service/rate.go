package service

import (
	"context"
	"fmt"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/rate"
)

type RateFetcher interface {
	FetchRate(ctx context.Context, ccFrom, ccTo string) (rate.Rate, error)
}

type RateService struct {
	fetchers []RateFetcher
}

var _ RateFetcher = (*RateService)(nil) // Ensure RateService implements RateFetcher

func (s *RateService) FetchRate(ctx context.Context, from, to string) (rate.Rate, error) {
	for _, f := range s.fetchers {
		r, err := f.FetchRate(ctx, from, to)
		if err != nil {
			fmt.Printf("failed to fetch rate: %v\n", err)
			continue
		}
		return r, nil
	}
	return rate.Rate{}, fmt.Errorf("failed to fetch rate")
}

func (s *RateService) SetNext(f ...RateFetcher) {
	s.fetchers = append(s.fetchers, f...)
}

func NewRateService(fetcher ...RateFetcher) *RateService {
	return &RateService{
		fetchers: fetcher,
	}
}

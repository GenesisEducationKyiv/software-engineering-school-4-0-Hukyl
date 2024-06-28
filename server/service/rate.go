package service

import (
	"context"
	"fmt"

	"github.com/Hukyl/genesis-kma-school-entry/models"
	"github.com/Hukyl/genesis-kma-school-entry/rate"
)

type RateRepo interface {
	Create(rate *models.Rate) error
}

type RateFetcher interface {
	FetchRate(ctx context.Context, ccFrom, ccTo string) (rate.Rate, error)
}

type RateService struct {
	repo    RateRepo
	fetcher RateFetcher
}

func (s *RateService) createRate(r rate.Rate) (*models.Rate, error) {
	row := &models.Rate{
		CurrencyFrom: r.CurrencyFrom,
		CurrencyTo:   r.CurrencyTo,
		Rate:         r.Rate,
	}
	err := s.repo.Create(row)
	if err != nil {
		return nil, fmt.Errorf("service rate creating: %w", err)
	}
	return row, nil
}

func (s *RateService) fetchRate(ctx context.Context, from, to string) (rate.Rate, error) {
	return s.fetcher.FetchRate(ctx, from, to)
}

func (s *RateService) FetchRate(ctx context.Context, from, to string) (*models.Rate, error) {
	r, err := s.fetchRate(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("service rate fetching: %w", err)
	}
	return s.createRate(r)
}

func NewRateService(repo RateRepo, fetcher RateFetcher) *RateService {
	return &RateService{
		repo:    repo,
		fetcher: fetcher,
	}
}

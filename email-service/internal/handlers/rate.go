package handlers

import (
	"context"
	"log/slog"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/models"
)

type rateRepository interface {
	Create(rate *models.Rate) error
}

func doWithContext(ctx context.Context, f func() error) error {
	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := f(); err != nil {
			slog.Error("error", slog.Any("error", err))
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-done:
			return nil
		}
	}
}

type RateEvents struct {
	repo rateRepository
}

func (s *RateEvents) SaveRate(
	ctx context.Context,
	from, to string, value float32, time time.Time,
) error {
	slog.Info(
		"rate fetched",
		slog.Any("from", from),
		slog.Any("to", to),
		slog.Any("rate", value),
		slog.Any("time", time),
	)
	rate := &models.Rate{
		CurrencyFrom: from,
		CurrencyTo:   to,
		Rate:         value,
		Created:      time.Unix(),
	}
	err := doWithContext(ctx, func() error {
		return s.repo.Create(rate)
	})
	if err != nil {
		slog.Error("saving rate", slog.Any("error", err))
		return err
	}
	return nil
}

func NewRateEvents(repo rateRepository) *RateEvents {
	return &RateEvents{repo: repo}
}

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
			getLogger().Error("error", slog.Any("error", err))
		}
		getLogger().Debug("function done", slog.Any("function", f))
	}()

	for {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			getLogger().Error("context done", slog.Any("error", err))
			return err
		case <-done:
			getLogger().Debug("done")
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
	getLogger().Info(
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
		getLogger().Error("saving rate", slog.Any("error", err))
		return err
	}
	getLogger().Debug("rate saved")
	return nil
}

func NewRateEvents(repo rateRepository) *RateEvents {
	return &RateEvents{repo: repo}
}

package handlers

import (
	"context"
	"log/slog"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/models"
)

var logger *slog.Logger

func getLogger() *slog.Logger {
	if logger == nil {
		logger = slog.Default().With(slog.String("src", "handler"))
	}
	return logger
}

type subscriberRepository interface {
	Create(subscriber *models.Subscriber) error
	Delete(subscriber *models.Subscriber) error
}

type SubscriberEvents struct {
	repo subscriberRepository
}

func (s *SubscriberEvents) Subscribe(ctx context.Context, email string) error {
	getLogger().Info("new subscriber")
	sub := &models.Subscriber{Email: email}
	err := doWithContext(ctx, func() error {
		getLogger().Info("saving subscriber")
		return s.repo.Create(sub)
	})
	if err != nil {
		getLogger().Error("saving subscriber", slog.Any("error", err))
	}
	getLogger().Debug("subscriber saved")
	return nil
}

func (s *SubscriberEvents) Unsubscribe(ctx context.Context, email string) error {
	slog.Info("delete subscriber")
	subscriber := &models.Subscriber{Email: email}
	err := doWithContext(ctx, func() error {
		slog.Info("deleting subscriber")
		return s.repo.Delete(subscriber)
	})
	if err != nil {
		slog.Error("deleting subscriber", slog.Any("error", err))
	}
	slog.Debug("subscriber deleted")
	return nil
}

func NewSubscriberEvents(subRepo subscriberRepository) *SubscriberEvents {
	return &SubscriberEvents{repo: subRepo}
}

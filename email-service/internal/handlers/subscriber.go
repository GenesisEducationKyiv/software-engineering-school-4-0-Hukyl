package handlers

import (
	"context"
	"log/slog"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/models"
)

type subscriberRepository interface {
	Create(subscriber *models.Subscriber) error
	Delete(subscriber *models.Subscriber) error
}

type SubscriberEvents struct {
	repo subscriberRepository
}

func (s *SubscriberEvents) Subscribe(ctx context.Context, email string) error {
	slog.Info("new subscriber", slog.Any("email", email))
	sub := &models.Subscriber{Email: email}
	err := doWithContext(ctx, func() error {
		return s.repo.Create(sub)
	})
	if err != nil {
		slog.Error("saving subscriber", slog.Any("error", err))
	}
	return nil
}

func (s *SubscriberEvents) Unsubscribe(ctx context.Context, email string) error {
	slog.Info("delete subscriber", slog.Any("email", email))
	subscriber := &models.Subscriber{Email: email}
	err := doWithContext(ctx, func() error {
		return s.repo.Delete(subscriber)
	})
	if err != nil {
		slog.Error("deleting subscriber", slog.Any("error", err))
	}
	return nil
}

func NewSubscriberEvents(subRepo subscriberRepository) *SubscriberEvents {
	return &SubscriberEvents{repo: subRepo}
}

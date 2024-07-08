package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/broker/rate"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/broker/subscriber"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/models"
	transportCfg "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker/transport/config"
)

const emailTimeout = 5 * time.Second

const (
	rateQueueName = "rate"
	userQueueName = "user"
)

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

func InitRateConsumer(config transportCfg.Config, rateRepo *models.RateRepository) *rate.Client {
	rateConsumer, err := rate.NewClient(transportCfg.Config{
		QueueName: rateQueueName,
		BrokerURI: config.BrokerURI,
	})
	if err != nil {
		slog.Error("creating rate client", slog.Any("error", err))
		return nil
	}
	defer rateConsumer.Close()
	rateConsumer.Subscribe(func( // nolint: errcheck
		ctx context.Context,
		from, to string, rate float32, time time.Time,
	) error {
		slog.Info(
			"rate fetched",
			slog.Any("from", from),
			slog.Any("to", to),
			slog.Any("rate", rate),
			slog.Any("time", time),
		)
		rateModel := &models.Rate{
			CurrencyFrom: from,
			CurrencyTo:   to,
			Rate:         rate,
			Created:      time.Unix(),
		}
		err := doWithContext(ctx, func() error {
			return rateRepo.Create(rateModel)
		})
		if err != nil {
			slog.Error("saving rate", slog.Any("error", err))
			return err
		}
		return nil
	})
	return rateConsumer
}

func InitSubscriberConsumer(
	config transportCfg.Config, subRepo *models.SubscriberRepository,
) *subscriber.Client {
	subConsumer, err := subscriber.NewClient(transportCfg.Config{
		QueueName: userQueueName,
		BrokerURI: config.BrokerURI,
	})
	if err != nil {
		slog.Error("creating subscriber client", slog.Any("error", err))
		return nil
	}
	defer subConsumer.Close()
	subConsumer.SubscribeCreate(func(ctx context.Context, email string) error { // nolint: errcheck
		slog.Info("new subscriber", slog.Any("email", email))
		sub := &models.Subscriber{Email: email}
		err := doWithContext(ctx, func() error {
			return subRepo.Create(sub)
		})
		if err != nil {
			slog.Error("saving subscriber", slog.Any("error", err))
		}
		return nil
	})
	subConsumer.SubscribeDelete(func(ctx context.Context, email string) error { // nolint: errcheck
		slog.Info("delete subscriber", slog.Any("email", email))
		subscriber := &models.Subscriber{Email: email}

		err := doWithContext(ctx, func() error {
			return subRepo.Delete(subscriber)
		})
		if err != nil {
			slog.Error("deleting subscriber", slog.Any("error", err))
		}
		return nil
	})
	return subConsumer
}

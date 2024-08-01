package handler

import (
	"context"
	"log/slog"
	"time"

	userBroker "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/broker/user"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/models"
)

const eventTimeout = 5 * time.Second

type UserRepo interface {
	Create(user *models.User) error
	Delete(user *models.User) error
	Exists(user *models.User) (bool, error)
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

type UserRepositorySaga struct {
	repo     UserRepo
	producer *userBroker.Producer
}

func (u *UserRepositorySaga) Create(user *models.User) error {
	if err := u.repo.Create(user); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), eventTimeout)
	defer cancel()
	if err := u.producer.SendSubscribe(ctx, user.Email); err != nil {
		return err
	}
	return nil
}

func (u *UserRepositorySaga) Delete(user *models.User) error {
	if err := u.repo.Delete(user); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), eventTimeout)
	defer cancel()
	if err := u.producer.SendUnsubscribe(ctx, user.Email); err != nil {
		return err
	}
	return nil
}

func (u *UserRepositorySaga) Exists(user *models.User) (bool, error) {
	// just a proxy method
	return u.repo.Exists(user)
}

func NewUserRepositorySaga(
	repo UserRepo,
	producer *userBroker.Producer,
	consumer *userBroker.Consumer,
) *UserRepositorySaga {
	urd := &UserRepositorySaga{
		repo:     repo,
		producer: producer,
	}
	err := consumer.ListenSubscribeCompensate(func(ctx context.Context, email string) error {
		err := doWithContext(ctx, func() error {
			return urd.repo.Delete(&models.User{Email: email})
		})
		if err != nil {
			slog.Error("subscribed compensate", slog.Any("error", err))
		}
		return nil
	})
	if err != nil {
		slog.Error("subscribing compensate", slog.Any("error", err))
	}
	slog.Debug("repository saga subscribed to subscribe compensate")
	err = consumer.ListenUnsubscribeCompensate(func(ctx context.Context, email string) error {
		err := doWithContext(ctx, func() error {
			return urd.repo.Create(&models.User{Email: email})
		})
		if err != nil {
			slog.Error("unsubscribed compensate", slog.Any("error", err))
		}
		return nil
	})
	if err != nil {
		slog.Error("unsubscribing compensate", slog.Any("error", err))
	}
	slog.Debug("repository saga subscribed to unsubscribe compensate")
	return urd
}

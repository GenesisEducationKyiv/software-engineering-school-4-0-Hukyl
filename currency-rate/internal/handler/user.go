package handler

import (
	"context"
	"log/slog"
	"time"

	userBroker "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/broker/user"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/models"
	"github.com/VictoriaMetrics/metrics"
)

const eventTimeout = 5 * time.Second

var logger *slog.Logger

var (
	sagaStepDuration = metrics.GetOrCreateHistogram(
		`saga_step_duration_seconds`,
	)
	createSagaCounter = metrics.GetOrCreateCounter(
		`saga_total{model="user", action="create"}`,
	)
	deleteSagaCounter = metrics.GetOrCreateCounter(
		`saga_total{model="user", action="delete"}`,
	)
	createCompensateCounter = metrics.GetOrCreateCounter(
		`compensate_total{model="user", action="create"}`,
	)
	deleteCompensateCounter = metrics.GetOrCreateCounter(
		`compensate_total{model="user", action="delete"}`,
	)
)

func getLogger() *slog.Logger {
	if logger == nil {
		logger = slog.Default().With(slog.Any("src", "userSagaHandler"))
	}
	return logger
}

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
			getLogger().Error("error", slog.Any("error", err))
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
	defer sagaStepDuration.UpdateDuration(time.Now())
	getLogger().Debug("creating user")
	if err := u.repo.Create(user); err != nil {
		return err
	}
	createSagaCounter.Inc()
	ctx, cancel := context.WithTimeout(context.Background(), eventTimeout)
	defer cancel()
	if err := u.producer.SendSubscribe(ctx, user.Email); err != nil {
		return err
	}
	getLogger().Debug("user created")
	return nil
}

func (u *UserRepositorySaga) Delete(user *models.User) error {
	defer sagaStepDuration.UpdateDuration(time.Now())
	getLogger().Debug("deleting user")
	if err := u.repo.Delete(user); err != nil {
		return err
	}
	deleteSagaCounter.Inc()
	ctx, cancel := context.WithTimeout(context.Background(), eventTimeout)
	defer cancel()
	if err := u.producer.SendUnsubscribe(ctx, user.Email); err != nil {
		return err
	}
	getLogger().Debug("user deleted")
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
		createCompensateCounter.Inc()
		err := doWithContext(ctx, func() error {
			return urd.repo.Delete(&models.User{Email: email})
		})
		if err != nil {
			getLogger().Error("subscribed compensate", slog.Any("error", err))
		}
		return nil
	})
	if err != nil {
		getLogger().Error("subscribing compensate", slog.Any("error", err))
	}
	getLogger().Debug("repository saga subscribed to subscribe compensate")
	err = consumer.ListenUnsubscribeCompensate(func(ctx context.Context, email string) error {
		deleteCompensateCounter.Inc()
		err := doWithContext(ctx, func() error {
			return urd.repo.Create(&models.User{Email: email})
		})
		if err != nil {
			getLogger().Error("unsubscribed compensate", slog.Any("error", err))
		}
		return nil
	})
	if err != nil {
		getLogger().Error("unsubscribing compensate", slog.Any("error", err))
	}
	getLogger().Debug("repository saga subscribed to unsubscribe compensate")
	return urd
}

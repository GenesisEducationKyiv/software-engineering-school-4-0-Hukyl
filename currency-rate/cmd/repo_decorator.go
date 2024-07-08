package main

import (
	"context"
	"time"

	userProducer "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/broker/user"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/models"
)

const eventTimeout = 5 * time.Second

type UserRepoDecorator struct {
	*models.UserRepository
	producer *userProducer.Producer
}

func (u *UserRepoDecorator) Create(user *models.User) error {
	if err := u.UserRepository.Create(user); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), eventTimeout)
	defer cancel()
	if err := u.producer.SendSubscribe(ctx, user.Email); err != nil {
		return err
	}
	return nil
}

func (u *UserRepoDecorator) Delete(user *models.User) error {
	if err := u.UserRepository.Delete(user); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), eventTimeout)
	defer cancel()
	if err := u.producer.SendUnsubscribe(ctx, user.Email); err != nil {
		return err
	}
	return nil
}

func NewUserRepoDecorator(
	repo *models.UserRepository, producer *userProducer.Producer,
) *UserRepoDecorator {
	return &UserRepoDecorator{
		UserRepository: repo,
		producer:       producer,
	}
}

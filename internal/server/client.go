package server

import (
	"context"
	"time"

	"github.com/Hukyl/genesis-kma-school-entry/internal/models"
	"github.com/Hukyl/genesis-kma-school-entry/internal/server/config"
)

const RateTimeout = 3 * time.Second

type RateService interface {
	FetchRate(ctx context.Context, from, to string) (*models.Rate, error)
}

type Client struct {
	Config      config.Config
	RateService RateService
	UserRepo    UserRepository
}

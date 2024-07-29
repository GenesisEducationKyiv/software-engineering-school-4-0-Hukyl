package server

import (
	"context"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/models"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/server/config"
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

package server

import (
	"github.com/Hukyl/genesis-kma-school-entry/models"
	"github.com/Hukyl/genesis-kma-school-entry/server/config"
)

type RateService interface {
	FetchRate(from, to string) (*models.Rate, error)
}

type Client struct {
	Config      config.Config
	RateService RateService
	UserRepo    UserRepository
}

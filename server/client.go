package server

import (
	"github.com/Hukyl/genesis-kma-school-entry/models"
	"github.com/Hukyl/genesis-kma-school-entry/rate"
	"github.com/Hukyl/genesis-kma-school-entry/server/config"
)

type RateFetcher interface {
	FetchRate(ccFrom, ccTo string) (rate.Rate, error)
}

type Client struct {
	Config      config.Config
	RateFetcher RateFetcher
	UserRepo    models.UserRepository
}

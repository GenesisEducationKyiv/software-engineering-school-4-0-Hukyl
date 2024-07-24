package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/models"
	"github.com/gin-gonic/gin"
)

var logger *slog.Logger

func getLogger() *slog.Logger {
	if logger == nil {
		logger = slog.Default().With(slog.Any("src", "api"))
	}
	return logger
}

const (
	RatePath        = "/rate"
	SubscribePath   = "/subscribe"
	UnsubscribePath = "/unsubscribe"
	ccFrom          = "USD"
	ccTo            = "UAH"
)

type UserRepository interface {
	Exists(user *models.User) (bool, error)
	Create(user *models.User) error
	Delete(user *models.User) error
}

// NewGetRateHandler is a handler that fetches the exchange rate between USD and UAH
// from a RateFetcher interface and returns it as a JSON response.
func NewGetRateHandler(rateService RateService, timeout time.Duration) func(*gin.Context) {
	return func(c *gin.Context) {
		getLogger().Info("new rate request")
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		rate, err := rateService.FetchRate(ctx, ccFrom, ccTo)
		if err != nil {
			getLogger().Error("fetching rate", slog.Any("error", err))
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}
		getLogger().Debug("rate fetched", slog.Any("rate", rate))
		c.JSON(http.StatusOK, rate.Rate)
		getLogger().Info("rate request processed")
	}
}

// NewSubscribeUserHandler is a handler that subscribes a user by email.
// The email is passed as a POST parameter and is required.
// If the user is already subscribed, returns a 409 Conflict status code.
// If the subscription is successful, returns a 200 OK status code.
func NewSubscribeUserHandler(repo UserRepository) func(*gin.Context) {
	return func(c *gin.Context) {
		getLogger().Info("new subscribe request")
		email := c.PostForm("email")
		if email == "" {
			getLogger().Debug("invalid request")
			c.JSON(http.StatusBadRequest, "email is required")
			return
		}
		user := &models.User{Email: email}
		exists, err := repo.Exists(user)
		if err != nil {
			getLogger().Error("checking user", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		if exists {
			getLogger().Debug("user already exists")
			c.JSON(http.StatusConflict, "")
			return
		}
		err = repo.Create(user)
		if err != nil {
			getLogger().Error("creating user", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		getLogger().Debug("user subscribed")
		c.JSON(http.StatusOK, "")
		getLogger().Info("subscribe request processed")
	}
}

func UnsubscribeUserHandler(repo UserRepository) func(*gin.Context) {
	return func(c *gin.Context) {
		getLogger().Info("new unsubscribe request")
		email := c.PostForm("email")
		if email == "" {
			getLogger().Debug("invalid request")
			c.JSON(http.StatusBadRequest, "email is required")
			return
		}
		user := &models.User{Email: email}
		exists, err := repo.Exists(user)
		if err != nil {
			getLogger().Error("checking user", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		if !exists {
			getLogger().Debug("user does not exist")
			c.JSON(http.StatusGone, "")
			return
		}
		err = repo.Delete(user)
		if err != nil {
			getLogger().Error("deleting user", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		getLogger().Debug("user unsubscribed")
		c.JSON(http.StatusOK, "")
		getLogger().Info("unsubscribe request processed")
	}
}

func NewEngine(client Client) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET(RatePath, NewGetRateHandler(client.RateService, RateTimeout))
	r.POST(SubscribePath, NewSubscribeUserHandler(client.UserRepo))
	r.POST(UnsubscribePath, UnsubscribeUserHandler(client.UserRepo))
	return r
}

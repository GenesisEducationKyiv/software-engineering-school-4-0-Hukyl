package server

import (
	"context"
	"net/http"

	"github.com/Hukyl/genesis-kma-school-entry/models"
	"github.com/Hukyl/genesis-kma-school-entry/settings"
	"github.com/gin-gonic/gin"
)

// wrapWithContext is a helper function that wraps a handler function
// with a context and returns a gin.HandlerFunc.
func wrapWithContext(ctx context.Context, fn func(context.Context, *gin.Context)) gin.HandlerFunc {
	return func(c *gin.Context) {
		fn(ctx, c)
	}
}

// GetRate is a handler that fetches the exchange rate between USD and UAH
// from the National Bank of Ukraine API and returns it as a JSON response.
func GetRate(ctx context.Context, c *gin.Context) {
	apiClient, ok := ctx.Value(settings.APIClientKey).(Client)
	if !ok {
		c.JSON(http.StatusInternalServerError, "")
		return
	}
	rateFetcher := apiClient.RateFetcher
	rate, err := rateFetcher.FetchRate("USD", "UAH")
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, rate.Rate)
}

// SubscribeUser is a handler that subscribes a user by email.
// The email is passed as a POST parameter and is required.
// If the user is already subscribed, returns a 409 Conflict status code.
// If the subscription is successful, returns a 200 OK status code.
func SubscribeUser(ctx context.Context, c *gin.Context) {
	email := c.PostForm("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, "email is required")
		return
	}
	apiClient, ok := ctx.Value(settings.APIClientKey).(Client)
	if !ok {
		c.JSON(http.StatusInternalServerError, "")
		return
	}
	repo := apiClient.UserRepo
	user := &models.User{Email: email}
	exists, err := repo.Exists(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	if exists {
		c.JSON(http.StatusConflict, "")
		return
	}
	err = repo.Create(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, "")
}

func NewEngine(ctx context.Context) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET(
		"/rate",
		wrapWithContext(ctx, GetRate),
	)
	r.POST(
		"/subscribe",
		wrapWithContext(ctx, SubscribeUser),
	)
	return r
}

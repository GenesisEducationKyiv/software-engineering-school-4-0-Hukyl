package server

import (
	"net/http"

	"github.com/Hukyl/genesis-kma-school-entry/models"
	"github.com/gin-gonic/gin"
)

type UserRepository interface {
	Exists(user *models.User) (bool, error)
	Create(user *models.User) error
}

// GetRateWrapper is a handler that fetches the exchange rate between USD and UAH
// from a RateFetcher interface and returns it as a JSON response.
func GetRateWrapper(rateFetcher RateFetcher) func(*gin.Context) {
	return func(c *gin.Context) {
		rate, err := rateFetcher.FetchRate("USD", "UAH")
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}
		c.JSON(http.StatusOK, rate.Rate)
	}
}

// SubscribeUser is a handler that subscribes a user by email.
// The email is passed as a POST parameter and is required.
// If the user is already subscribed, returns a 409 Conflict status code.
// If the subscription is successful, returns a 200 OK status code.
func SubscribeUserWrapper(repo UserRepository) func(*gin.Context) {
	return func(c *gin.Context) {
		email := c.PostForm("email")
		if email == "" {
			c.JSON(http.StatusBadRequest, "email is required")
			return
		}
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
}

func NewEngine(client Client) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/rate", GetRateWrapper(client.RateFetcher))
	r.POST("/subscribe", SubscribeUserWrapper(&client.UserRepo))
	return r
}

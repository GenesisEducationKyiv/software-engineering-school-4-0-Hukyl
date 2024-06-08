package server

import (
	"net/http"

	"github.com/Hukyl/genesis-kma-school-entry/models"
	"github.com/Hukyl/genesis-kma-school-entry/rate"
	"github.com/gin-gonic/gin"
)

// GetRate is a handler that fetches the exchange rate between USD and UAH
// from the National Bank of Ukraine API and returns it as a JSON response.
func GetRate(c *gin.Context) {
	nbu := rate.NewNBURateFetcher()
	rate, err := nbu.FetchRate("USD", "UAH")
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
func SubscribeUser(c *gin.Context) {
	email := c.PostForm("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, "email is required")
		return
	}
	db := models.NewDB()
	exists, err := models.UserExists(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	if exists {
		c.JSON(http.StatusConflict, "")
		return
	}
	user := models.User{Email: email}
	err = db.Connection().Create(&user).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, "")
}

func ApiEngine() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/rate", GetRate)
	r.POST("/subscribe", SubscribeUser)
	return r
}

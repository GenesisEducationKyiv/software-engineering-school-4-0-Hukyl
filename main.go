package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Hukyl/genesis-kma-school-entry/models"
	"github.com/Hukyl/genesis-kma-school-entry/rate"
	"github.com/Hukyl/genesis-kma-school-entry/server"
	"github.com/Hukyl/genesis-kma-school-entry/settings"
	"github.com/Hukyl/genesis-kma-school-entry/utils"
	"github.com/gin-gonic/gin"
)

func NotifyUsers() {
	rate, err := rate.NewNBURateFetcher().FetchRate("USD", "UAH")
	if err != nil {
		log.Println(err)
		return
	}
	db := models.NewDB()
	var users []models.User
	db.Connection().Find(&users)
	log.Printf("Notifying by email %d users\n", len(users))
	for _, user := range users {
		message := fmt.Sprintf("1 USD = %f UAH", rate.Rate)
		if err := utils.SendEmail(user.Email, message); err != nil {
			log.Println(err)
		}
	}
}

func main() {
	settings.InitSettings()
	models.NewDB().Migrate()

	go func() {
		for {
			go NotifyUsers()
			time.Sleep(settings.EMAIL_INTERVAL)
		}
	}()

	router := server.ApiEngine()
	router.Use(gin.Logger())
	s := &http.Server{
		Addr:         ":" + settings.Port,
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Printf("Starting server on %s\n", s.Addr)
	s.ListenAndServe()
}

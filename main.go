package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/Hukyl/genesis-kma-school-entry/mail"
	mailCfg "github.com/Hukyl/genesis-kma-school-entry/mail/config"
	"github.com/Hukyl/genesis-kma-school-entry/models"
	modelsCfg "github.com/Hukyl/genesis-kma-school-entry/models/config"
	"github.com/Hukyl/genesis-kma-school-entry/rate"
	"github.com/Hukyl/genesis-kma-school-entry/server"
	serverCfg "github.com/Hukyl/genesis-kma-school-entry/server/config"
	"github.com/Hukyl/genesis-kma-school-entry/settings"
)

const emailInterval = 24 * time.Hour

func NotifyUsers(ctx context.Context, apiClient server.Client, mc *mail.Client) {
	db := apiClient.DB

	rate, err := apiClient.RateFetcher.FetchRate("USD", "UAH")
	if err != nil {
		slog.Warn(err.Error())
		return
	}
	var users []models.User
	db.Connection().Find(&users)
	slog.Info(fmt.Sprintf(
		"Notifying by email %d users\n",
		len(users),
	))
	for _, user := range users {
		message := fmt.Sprintf("1 USD = %f UAH", rate.Rate)
		if err := mc.SendEmail(ctx, user.Email, message); err != nil {
			slog.Error(err.Error())
		}
	}
}

func main() {
	err := settings.InitSettings()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to initialize settings: %s", err))
		panic(err)
	}

	ctx := context.Background()
	ctx = context.WithValue(
		ctx,
		settings.DebugKey,
		os.Getenv("DEBUG") == "true",
	)

	mc := &mail.Client{Config: mailCfg.NewFromEnv()}
	apiClient := server.Client{
		Config:      serverCfg.NewFromEnv(),
		RateFetcher: rate.NewNBURateFetcher(),
		DB:          models.NewDB(modelsCfg.NewFromEnv()),
	}
	ctx = context.WithValue(ctx, settings.APIClientKey, apiClient)
	apiClient.Engine = server.NewEngine(ctx)

	err = apiClient.DB.Migrate()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to migrate database: %s", err))
		panic(err)
	}

	go func() {
		for {
			go NotifyUsers(ctx, apiClient, mc)
			time.Sleep(emailInterval)
		}
	}()

	s := server.NewServer(apiClient.Config, apiClient.Engine)
	if err = s.ListenAndServe(); err != nil {
		slog.Error(fmt.Sprintf("HTTP server error occurred: %s", err))
	}
}

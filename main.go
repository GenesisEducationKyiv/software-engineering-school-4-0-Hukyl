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
	slog.Info(
		"notifying users by email",
		slog.Any("userCount", len(users)),
	)
	for _, user := range users {
		message := fmt.Sprintf("1 USD = %f UAH", rate.Rate)
		if err := mc.SendEmail(ctx, user.Email, message); err != nil {
			slog.Error(err.Error())
		}
	}
}

func main() {
	if err := settings.InitSettings(); err != nil {
		slog.Error("failed to initialize settings", slog.Any("error", err))
		panic(err)
	}

	ctx := context.Background()
	ctx = context.WithValue(
		ctx,
		settings.DebugKey,
		os.Getenv("DEBUG") == "true",
	)

	mc := &mail.Client{Config: mailCfg.NewFromEnv()}
	db, err := models.NewDB(modelsCfg.NewFromEnv())
	if err != nil {
		slog.Error("failed to initialize database", slog.Any("error", err))
		panic(err)
	}
	apiClient := server.Client{
		Config:      serverCfg.NewFromEnv(),
		RateFetcher: rate.NewNBURateFetcher(),
		DB:          db,
	}
	ctx = context.WithValue(ctx, settings.APIClientKey, apiClient)
	apiClient.Engine = server.NewEngine(ctx)

	if err := apiClient.DB.Migrate(); err != nil {
		slog.Error("failed to migrate database", slog.Any("error", err))
		panic(err)
	}

	go func() {
		for {
			go NotifyUsers(ctx, apiClient, mc)
			time.Sleep(emailInterval)
		}
	}()

	s := server.NewServer(apiClient.Config, apiClient.Engine)
	if err := s.ListenAndServe(); err != nil {
		slog.Error("HTTP server error occurred", slog.Any("error", err))
	}
}

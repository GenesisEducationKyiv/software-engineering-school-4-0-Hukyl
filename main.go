package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/Hukyl/genesis-kma-school-entry/database"
	dbCfg "github.com/Hukyl/genesis-kma-school-entry/database/config"
	"github.com/Hukyl/genesis-kma-school-entry/mail"
	mailCfg "github.com/Hukyl/genesis-kma-school-entry/mail/config"
	"github.com/Hukyl/genesis-kma-school-entry/models"
	"github.com/Hukyl/genesis-kma-school-entry/rate"
	"github.com/Hukyl/genesis-kma-school-entry/server"
	serverCfg "github.com/Hukyl/genesis-kma-school-entry/server/config"
	"github.com/Hukyl/genesis-kma-school-entry/server/notifications"
	"github.com/Hukyl/genesis-kma-school-entry/settings"
	"github.com/robfig/cron/v3"
)

const defaultCronSpec = "0 0 12 * * *"

func StartCron(spec string, f func()) *cron.Cron {
	c := cron.New()
	_, err := c.AddFunc(spec, f)
	if err != nil {
		slog.Error("failed to add cron job", slog.Any("error", err))
		return nil
	}
	slog.Info("cron job added", slog.Any("spec", spec))
	c.Start()
	return c
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
	db, err := database.New(dbCfg.NewFromEnv())
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

	if err := apiClient.DB.Migrate(&models.User{}); err != nil {
		slog.Error("failed to migrate database", slog.Any("error", err))
		panic(err)
	}

	cronSpec := os.Getenv("CRON_SPEC")
	if cronSpec == "" {
		cronSpec = defaultCronSpec
	}
	notifier := notifications.NewUsersNotifier(
		mc,
		apiClient.RateFetcher,
		models.NewUserRepository(apiClient.DB),
	)
	StartCron(cronSpec, func() {
		notifier.Notify(ctx)
	})

	s := server.NewServer(apiClient.Config, apiClient.Engine)
	if err := s.ListenAndServe(); err != nil {
		slog.Error("HTTP server error occurred", slog.Any("error", err))
	}
}

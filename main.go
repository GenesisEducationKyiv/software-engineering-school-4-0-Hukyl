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
	"github.com/Hukyl/genesis-kma-school-entry/server/notifications/message"
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

	// Initialize database
	db, err := database.New(dbCfg.NewFromEnv())
	if err != nil {
		slog.Error("failed to initialize database", slog.Any("error", err))
		panic(err)
	}
	if err := db.Migrate(&models.User{}); err != nil {
		slog.Error("failed to migrate database", slog.Any("error", err))
		panic(err)
	}

	apiClient := server.Client{
		Config:      serverCfg.NewFromEnv(),
		RateFetcher: rate.NewNBURateFetcher(),
		UserRepo:    *models.NewUserRepository(db),
	}

	// Start cron job for notifications
	cronSpec := os.Getenv("CRON_SPEC")
	if cronSpec == "" {
		slog.Warn(
			"CRON_SPEC is not set, using default value",
			slog.Any("default", defaultCronSpec),
		)
		cronSpec = defaultCronSpec
	}
	notifier := notifications.NewUsersNotifier(
		&mail.Client{Config: mailCfg.NewFromEnv()},
		apiClient.RateFetcher,
		&apiClient.UserRepo,
		&message.PlainRateMessage{},
	)
	StartCron(cronSpec, func() {
		notifier.Notify(ctx)
	})

	// Start HTTP server
	s := server.NewServer(apiClient.Config, server.NewEngine(apiClient))
	slog.Info("starting server", slog.Any("address", s.Addr))
	if err := s.ListenAndServe(); err != nil {
		slog.Error("HTTP server error occurred", slog.Any("error", err))
	}
}

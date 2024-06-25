package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/Hukyl/genesis-kma-school-entry/database"
	dbCfg "github.com/Hukyl/genesis-kma-school-entry/database/config"
	"github.com/Hukyl/genesis-kma-school-entry/mail"
	"github.com/Hukyl/genesis-kma-school-entry/mail/backends"
	mailCfg "github.com/Hukyl/genesis-kma-school-entry/mail/config"
	"github.com/Hukyl/genesis-kma-school-entry/models"
	"github.com/Hukyl/genesis-kma-school-entry/rate/fetchers"
	"github.com/Hukyl/genesis-kma-school-entry/server"
	serverCfg "github.com/Hukyl/genesis-kma-school-entry/server/config"
	"github.com/Hukyl/genesis-kma-school-entry/server/notifications"
	"github.com/Hukyl/genesis-kma-school-entry/server/notifications/message"
	"github.com/Hukyl/genesis-kma-school-entry/server/service"
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

func InitDatabase() (*database.DB, error) {
	db, err := database.New(dbCfg.NewFromEnv())
	if err != nil {
		return nil, err
	}
	if err := db.Migrate(&models.User{}, &models.Rate{}); err != nil {
		return nil, err
	}
	return db, nil
}

func InitFetchers() fetchers.RateFetcher {
	// Initialize rate fetcher chain of responsibilities
	baseFetcher := fetchers.NewBaseFetcher()
	nbuFetcher := fetchers.NewNBURateFetcher()
	currencyBeaconFetcher := fetchers.NewCurrencyBeaconFetcher(
		os.Getenv("CURRENCY_BEACON_API_KEY"),
	)
	nbuFetcher.SetNext(baseFetcher)
	currencyBeaconFetcher.SetNext(nbuFetcher)
	return currencyBeaconFetcher
}

func main() {
	if err := settings.InitSettings(); err != nil {
		slog.Error("failed to initialize settings", slog.Any("error", err))
		panic(err)
	}

	debug := os.Getenv("DEBUG") == "true"

	db, err := InitDatabase()
	if err != nil {
		slog.Error("failed to initialize database", slog.Any("error", err))
		panic(err)
	}

	// Initialize rate fetcher chain of responsibilities
	rateFetcher := InitFetchers()

	userRepo := models.NewUserRepository(db)
	apiClient := server.Client{
		Config:      serverCfg.NewFromEnv(),
		RateService: service.NewRateService(models.NewRateRepository(db), rateFetcher),
		UserRepo:    userRepo,
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
	var mailer mail.Mailer
	mailConfig := mailCfg.NewFromEnv()
	if debug {
		mailer = backends.NewConsoleMailer(mailConfig)
	} else {
		mailer = backends.NewGomailMailer(mailConfig)
	}
	notifier := notifications.NewUsersNotifier(
		mail.NewClient(mailer),
		apiClient.RateService,
		userRepo,
		&message.PlainRate{},
	)
	StartCron(cronSpec, func() {
		ctx, cancel := context.WithTimeout(context.Background(), server.RateTimeout)
		defer cancel()
		notifier.Notify(ctx)
	})

	// Start HTTP server
	s := server.NewServer(apiClient.Config, server.NewEngine(apiClient))
	slog.Info("starting server", slog.Any("address", s.Addr))
	if err := s.ListenAndServe(); err != nil {
		slog.Error("HTTP server error occurred", slog.Any("error", err))
	}
}

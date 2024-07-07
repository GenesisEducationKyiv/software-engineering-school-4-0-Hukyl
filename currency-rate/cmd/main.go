package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/database"
	dbCfg "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/database/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/mail"
	transportCfg "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/mail/transport/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/models"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/rate/fetchers"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/server"
	serverCfg "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/server/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/server/notifications"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/server/notifications/message"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/server/service"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/settings"
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
	nbuFetcher := fetchers.NewNBURateFetcher()
	currencyBeaconFetcher := fetchers.NewCurrencyBeaconFetcher(
		os.Getenv("CURRENCY_BEACON_API_KEY"),
	)
	currencyBeaconFetcher.SetNext(nbuFetcher)
	return currencyBeaconFetcher
}

func main() {
	if err := settings.InitSettings(); err != nil {
		slog.Error("failed to initialize settings", slog.Any("error", err))
		// panic(err)
	}

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

	mailerFacade, err := mail.NewMailerFacade(transportCfg.NewFromEnv())
	if err != nil {
		slog.Error("failed to initialize mailer facade", slog.Any("error", err))
	}
	defer mailerFacade.Close()
	notifier := notifications.NewUsersNotifier(
		mailerFacade,
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

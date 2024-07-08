package main

import (
	"log/slog"
	"os"

	rateProducer "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/broker/rate"
	userProducer "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/broker/user"
	cronRate "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/cron/rate"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/models"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/rate/fetchers"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/server"
	serverCfg "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/server/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/server/service"
	transportCfg "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker/transport/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/cron"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/database"
	dbCfg "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/database/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/settings"
)

const (
	ccFrom = "USD"
	ccTo   = "UAH"
)

const (
	rateQueueName = "rate"
	userQueueName = "user"
)

func InitDatabase() (*database.DB, error) {
	db, err := database.New(dbCfg.NewFromEnv())
	if err != nil {
		return nil, err
	}
	if err := db.Migrate(&models.User{}); err != nil {
		return nil, err
	}
	return db, nil
}

func InitFetchers() []service.RateFetcher {
	nbuFetcher := fetchers.NewNBURateFetcher()
	currencyBeaconFetcher := fetchers.NewCurrencyBeaconFetcher(
		os.Getenv("CURRENCY_BEACON_API_KEY"),
	)
	return []service.RateFetcher{currencyBeaconFetcher, nbuFetcher}
}

func InitCron(fetcher service.RateFetcher) *cron.Manager {
	cfg := transportCfg.NewFromEnv()
	cfg.QueueName = rateQueueName

	producer, err := rateProducer.NewProducer(cfg)
	if err != nil {
		slog.Error("failed to initialize rate producer", slog.Any("error", err))
		return nil
	}

	job := cronRate.NewCronJob(fetcher, producer, ccFrom, ccTo)
	spec := "@every 5m"

	cronManager := cron.NewManager()
	cronManager.AddJob(spec, job.Run) // nolint: errcheck
	return cronManager
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
	rateService := service.NewRateService(InitFetchers()...)

	// Initializer user event producer
	userProducer, err := userProducer.NewProducer(transportCfg.NewFromEnv())
	if err != nil {
		slog.Error("failed to initialize user producer", slog.Any("error", err))
		panic(err)
	}
	userRepo := models.NewUserRepository(db)

	// Decorate userRepo with userProducer
	decoratedUserRepo := NewUserRepoDecorator(userRepo, userProducer)

	apiClient := server.Client{
		Config:      serverCfg.NewFromEnv(),
		RateService: rateService,
		UserRepo:    decoratedUserRepo,
	}

	cronManager := InitCron(rateService)
	if cronManager == nil {
		slog.Error("failed to initialize cron manager")
	} else {
		cronManager.Start()
		defer cronManager.Stop()
	}

	// Start HTTP server
	s := server.NewServer(apiClient.Config, server.NewEngine(apiClient))
	slog.Info("starting server", slog.Any("address", s.Addr))
	if err := s.ListenAndServe(); err != nil {
		slog.Error("HTTP server error occurred", slog.Any("error", err))
	}
}

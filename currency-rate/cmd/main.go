package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	rateProducer "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/broker/rate"
	userProducer "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/broker/user"
	appCfg "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/config"
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

var appConfig appCfg.Config

const eventTimeout = 5 * time.Second

type UserRepoDecorator struct {
	*models.UserRepository
	producer *userProducer.Producer
}

func (u *UserRepoDecorator) Create(user *models.User) error {
	if err := u.UserRepository.Create(user); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), eventTimeout)
	defer cancel()
	if err := u.producer.SendSubscribe(ctx, user.Email); err != nil {
		return err
	}
	return nil
}

func (u *UserRepoDecorator) Delete(user *models.User) error {
	if err := u.UserRepository.Delete(user); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), eventTimeout)
	defer cancel()
	if err := u.producer.SendUnsubscribe(ctx, user.Email); err != nil {
		return err
	}
	return nil
}

func NewDatabase() (*database.DB, error) {
	db, err := database.New(dbCfg.NewFromEnv())
	if err != nil {
		return nil, err
	}
	if err := db.Migrate(&models.User{}); err != nil {
		return nil, err
	}
	return db, nil
}

func NewLogger() *slog.Logger {
	loggerOptions := &slog.HandlerOptions{}
	if appConfig.Debug {
		loggerOptions.Level = slog.LevelDebug
	}
	logger := slog.New(
		slog.NewTextHandler(os.Stdout, loggerOptions),
	).With(slog.Any("service", "currency-rate"))
	return logger
}

func NewFetchers() []service.RateFetcher {
	nbuFetcher := fetchers.NewNBURateFetcher()
	currencyBeaconFetcher := fetchers.NewCurrencyBeaconFetcher(
		appConfig.CurrencyBeaconAPIKey,
	)
	return []service.RateFetcher{currencyBeaconFetcher, nbuFetcher}
}

func NewCron(
	transportConfig transportCfg.Config,
	fetcher service.RateFetcher,
) (*cron.Manager, error) {
	transportConfig.QueueName = appConfig.RateQueueName

	producer, err := rateProducer.NewProducer(transportConfig)
	if err != nil {
		slog.Error("failed to create rate producer", slog.Any("error", err))
		return nil, err
	}

	job := cronRate.NewCronJob(fetcher, producer, ccFrom, ccTo)
	spec := appConfig.RateRefreshCropSpec
	if spec == "" {
		spec = "@every 5m"
	}

	cronManager := cron.NewManager()
	err = cronManager.AddJob(spec, job.Run)
	if err != nil {
		slog.Error("adding cron job", slog.Any("error", err))
		return nil, err
	}
	return cronManager, nil
}

func main() {
	if err := settings.InitSettings(); err != nil {
		slog.Error("failed to initialize settings", slog.Any("error", err))
		// panic(err)
	}

	appConfig = appCfg.NewFromEnv()
	slog.SetDefault(NewLogger())

	db, err := NewDatabase()
	if err != nil {
		slog.Error("failed to create database", slog.Any("error", err))
		panic(err)
	}

	// Initialize rate fetcher chain of responsibilities
	rateService := service.NewRateService(NewFetchers()...)

	transportConfig := transportCfg.NewFromEnv()

	// Initializer user event producer
	userProducer, err := userProducer.NewProducer(transportCfg.Config{
		BrokerURI: transportConfig.BrokerURI,
		QueueName: appConfig.UserQueueName,
	})
	if err != nil {
		slog.Error("failed to create user producer", slog.Any("error", err))
		panic(err)
	}
	userRepo := models.NewUserRepository(db)

	// Decorate userRepo with userProducer
	decoratedUserRepo := &UserRepoDecorator{
		UserRepository: userRepo,
		producer:       userProducer,
	}

	apiClient := server.Client{
		Config:      serverCfg.NewFromEnv(),
		RateService: rateService,
		UserRepo:    decoratedUserRepo,
	}

	cronManager, err := NewCron(transportConfig, rateService)
	if err != nil {
		slog.Error("failed to initialize cron manager", slog.Any("error", err))
		panic(err)
	}
	cronManager.Start()
	defer cronManager.Stop()

	// Start HTTP server
	s := server.NewServer(apiClient.Config, server.NewEngine(apiClient))
	slog.Info("starting server", slog.Any("address", s.Addr))
	if err := s.ListenAndServe(); err != nil {
		slog.Error("HTTP server error occurred", slog.Any("error", err))
	}
}

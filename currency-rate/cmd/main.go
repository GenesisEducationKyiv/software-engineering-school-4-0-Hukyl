package main

import (
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"time"

	rateProducer "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/broker/rate"
	userBroker "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/broker/user"
	appCfg "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/config"
	cronRate "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/cron/rate"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/handler"
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
	"github.com/VictoriaMetrics/metrics"
)

const (
	ccFrom = "USD"
	ccTo   = "UAH"
)

var appConfig appCfg.Config

func NewDatabase() (*database.DB, error) {
	db, err := database.New(dbCfg.NewFromEnv())
	if err != nil {
		return nil, err
	}
	if err := db.Migrate(&models.User{}); err != nil {
		return nil, err
	}
	slog.Debug("database created")
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
	fetchers := []service.RateFetcher{currencyBeaconFetcher, nbuFetcher}
	slog.Debug("fetchers initialized", slog.Any("fetchers", fetchers))
	return fetchers
}

func NewCron(transportConfig transportCfg.Config, fetcher service.RateFetcher) *cron.Manager {
	transportConfig.QueueName = appConfig.RateQueueName

	producer, err := rateProducer.NewProducer(transportConfig)
	if err != nil {
		slog.Error("failed to create rate producer", slog.Any("error", err))
		return nil
	}
	slog.Debug("rate producer created")

	job := cronRate.NewCronJob(fetcher, producer, ccFrom, ccTo)
	spec := appConfig.RateRefreshCronSpec
	if spec == "" {
		spec = "@every 5m"
	}
	slog.Debug("cron job spec", slog.Any("spec", spec))

	cronManager := cron.NewManager()
	err = cronManager.AddJob(spec, job.Run)
	if err != nil {
		slog.Error("adding cron job", slog.Any("error", err))
		return nil
	}
	return cronManager
}

// Get preferred outbound ip of this machine.
func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	return conn.LocalAddr().String()
}

func main() { // nolint: funlen
	if err := settings.InitSettings(); err != nil {
		slog.Error("failed to initialize settings", slog.Any("error", err))
	}
	appConfig = appCfg.NewFromEnv()
	slog.SetDefault(NewLogger())
	slog.Debug("settings initialized")

	db, err := NewDatabase()
	if err != nil {
		slog.Error("failed to create database", slog.Any("error", err))
		panic(err)
	}
	defer db.Close()

	// Initialize rate fetcher chain of responsibilities
	rateService := service.NewRateService(NewFetchers()...)

	transportConfig := transportCfg.NewFromEnv()
	// Initialize user event producer
	userProducer, err := userBroker.NewProducer(transportCfg.Config{
		BrokerURI: transportConfig.BrokerURI,
		QueueName: appConfig.UserQueueName,
	})
	if err != nil {
		slog.Error("failed to create user producer", slog.Any("error", err))
		panic(err)
	}
	userConsumer, err := userBroker.NewConsumer(transportCfg.Config{
		BrokerURI: transportConfig.BrokerURI,
		QueueName: appConfig.UserCompensateQueueName,
	})
	if err != nil {
		slog.Error("failed to create user consumer", slog.Any("error", err))
		panic(err)
	}
	go userConsumer.Start()
	userRepo := models.NewUserRepository(db)

	// Decorate userRepo with userProducer
	decoratedUserRepo := handler.NewUserRepositorySaga(userRepo, userProducer, userConsumer)
	slog.Debug("user repository saga created")

	apiClient := server.Client{
		Config:      serverCfg.NewFromEnv(),
		RateService: rateService,
		UserRepo:    decoratedUserRepo,
	}

	cronManager := NewCron(transportConfig, rateService)
	if cronManager == nil {
		slog.Error("failed to initialize cron manager")
	} else {
		cronManager.Start()
		defer cronManager.Stop()
	}

	// Start HTTP server
	s := server.NewServer(apiClient.Config, server.NewEngine(apiClient))
	slog.Info("starting server", slog.Any("address", s.Addr))

	err = metrics.InitPush(
		appConfig.VictoriaMetricsPushURL,
		10*time.Second,
		fmt.Sprintf(`job="currency-rate",instance="%s"`, getOutboundIP()),
		true,
	)
	if err != nil {
		slog.Error("failed to initialize metrics push", slog.Any("error", err))
	}

	if err := s.ListenAndServe(); err != nil {
		slog.Error("HTTP server error occurred", slog.Any("error", err))
	}
	slog.Debug("server stopped gracefully")
}

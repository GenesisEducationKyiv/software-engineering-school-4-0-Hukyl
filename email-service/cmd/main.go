package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/broker/rate"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/broker/subscriber"
	appCfg "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/handlers"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/mail"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/mail/backends"
	mailCfg "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/mail/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/models"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/notifications"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/notifications/message"
	transportCfg "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker/transport/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/cron"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/database"
	dbCfg "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/database/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/settings"
)

const defaultCronSpec = "0 0 12 * * *"

const emailTimeout = 5 * time.Second

var appConfig appCfg.Config

func NewDatabase() (*database.DB, error) {
	db, err := database.New(dbCfg.NewFromEnv())
	if err != nil {
		return nil, err
	}
	if err := db.Migrate(&models.Rate{}, &models.Subscriber{}); err != nil {
		return nil, err
	}
	return db, nil
}

func NewNotificationsCron(db *database.DB, mailer notifications.EmailClient) *cron.Manager {
	// Start cron job for notifications
	spec := appConfig.NotificationCropSpec
	if spec == "" {
		spec = defaultCronSpec
	}

	notifier := notifications.NewMailNotifier(
		mailer,
		models.NewRateRepository(db),
		models.NewSubscriberRepository(db),
		&message.PlainRate{},
	)

	notifyF := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), emailTimeout)
		defer cancel()
		return notifier.Notify(ctx)
	}

	cronManager := cron.NewManager()
	err := cronManager.AddJob(spec, notifyF)
	if err != nil {
		slog.Error("adding job", slog.Any("error", err))
	}
	return cronManager
}

func NewRateConsumer(config transportCfg.Config, rateRepo *models.RateRepository) *rate.Client {
	rateConsumer, err := rate.NewClient(transportCfg.Config{
		QueueName: appConfig.RateQueueName,
		BrokerURI: config.BrokerURI,
	})
	if err != nil {
		slog.Error("creating rate client", slog.Any("error", err))
		return nil
	}
	eventHandler := handlers.NewRateEvents(rateRepo)
	err = rateConsumer.Subscribe(eventHandler.SaveRate)
	if err != nil {
		slog.Error("subscribing to rate", slog.Any("error", err))
	}
	return rateConsumer
}

func NewSubscriberConsumer(
	config transportCfg.Config, subRepo *models.SubscriberRepository,
) *subscriber.CompensateClient {
	eventHandler := handlers.NewSubscriberEvents(subRepo)

	subConsumer, err := subscriber.NewCompensateClient(
		transportCfg.Config{
			QueueName: appConfig.UserQueueName,
			BrokerURI: config.BrokerURI,
		},
		transportCfg.Config{
			QueueName: appConfig.UserCompensateQueueName,
			BrokerURI: config.BrokerURI,
		},
		eventHandler.Subscribe,
		eventHandler.Unsubscribe,
	)
	if err != nil {
		slog.Error("creating subscriber client", slog.Any("error", err))
		return nil
	}

	return subConsumer
}

func NewLogger() *slog.Logger {
	loggerOptions := &slog.HandlerOptions{}
	if appConfig.Debug {
		loggerOptions.Level = slog.LevelDebug
	}
	logger := slog.New(
		slog.NewTextHandler(os.Stdout, loggerOptions),
	).With(slog.Any("service", "email-service"))
	return logger
}

func main() {
	// Initialize settings
	err := settings.InitSettings()
	if err != nil {
		slog.Error("initializing settings", slog.Any("error", err))
	}

	// Initialize app config
	appConfig = appCfg.NewFromEnv()
	slog.SetDefault(NewLogger())

	// Initialize database
	db, err := NewDatabase()
	if err != nil {
		slog.Error("creating database", slog.Any("error", err))
		panic(err)
	}
	rateRepo := models.NewRateRepository(db)
	subRepo := models.NewSubscriberRepository(db)

	// Initialize mailer by debug mode
	var mailer mail.Mailer
	mailConfig := mailCfg.NewFromEnv()
	if appConfig.Debug {
		mailer = backends.NewConsoleMailer(mailConfig)
	} else {
		mailer = backends.NewGomailMailer(mailConfig)
	}
	mailClient := mail.NewClient(mailer)

	// Initialize cron manager for notifications
	cronManager := NewNotificationsCron(db, mailClient)
	cronManager.Start()
	defer cronManager.Stop()

	transportConfig := transportCfg.NewFromEnv()
	// Create rate consumer and subscribe to events
	rateConsumer := NewRateConsumer(transportConfig, rateRepo)
	if rateConsumer == nil {
		slog.Error("initializing rate consumer")
	}
	defer rateConsumer.Close()

	// Create subscribe consumer and subscribe to events
	subConsumer := NewSubscriberConsumer(transportConfig, subRepo)
	if subConsumer == nil {
		slog.Error("initializing subscriber consumer")
	}
	defer subConsumer.Close()

	termChannel := make(chan os.Signal, 1)
	signal.Notify(termChannel, syscall.SIGINT)
	signal.Notify(termChannel, syscall.SIGTERM)
	<-termChannel
	// Gracefully close the client
}

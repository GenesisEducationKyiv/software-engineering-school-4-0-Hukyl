package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/broker/rate"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/broker/subscriber"
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

var (
	rateQueueName = os.Getenv("RATE_QUEUE_NAME")
	userQueueName = os.Getenv("USER_QUEUE_NAME")
)

func InitDatabase() (*database.DB, error) {
	db, err := database.New(dbCfg.NewFromEnv())
	if err != nil {
		return nil, err
	}
	if err := db.Migrate(&models.Rate{}, &models.Subscriber{}); err != nil {
		return nil, err
	}
	return db, nil
}

func InitNotificationsCron(db *database.DB, mailer notifications.EmailClient) *cron.Manager {
	// Start cron job for notifications
	envVar := "NOTIFICATION_CRON_SPEC"
	cronSpec := os.Getenv(envVar)
	if cronSpec == "" {
		slog.Warn(
			fmt.Sprintf("%s is not set, using default value", envVar),
			slog.Any("default", defaultCronSpec),
		)
		cronSpec = defaultCronSpec
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
	cronManager.AddJob(cronSpec, notifyF) // nolint: errcheck
	return cronManager
}

func doWithContext(ctx context.Context, f func() error) error {
	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := f(); err != nil {
			slog.Error("error", slog.Any("error", err))
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-done:
			return nil
		}
	}
}

func InitRateConsumer(config transportCfg.Config, rateRepo *models.RateRepository) *rate.Client {
	rateConsumer, err := rate.NewClient(transportCfg.Config{
		QueueName: rateQueueName,
		BrokerURI: config.BrokerURI,
	})
	if err != nil {
		slog.Error("creating rate client", slog.Any("error", err))
		return nil
	}
	rateConsumer.Subscribe(func( // nolint: errcheck
		ctx context.Context,
		from, to string, rate float32, time time.Time,
	) error {
		slog.Info(
			"rate fetched",
			slog.Any("from", from),
			slog.Any("to", to),
			slog.Any("rate", rate),
			slog.Any("time", time),
		)
		rateModel := &models.Rate{
			CurrencyFrom: from,
			CurrencyTo:   to,
			Rate:         rate,
			Created:      time.Unix(),
		}
		err := doWithContext(ctx, func() error {
			return rateRepo.Create(rateModel)
		})
		if err != nil {
			slog.Error("saving rate", slog.Any("error", err))
			return err
		}
		return nil
	})
	return rateConsumer
}

func InitSubscriberConsumer(
	config transportCfg.Config, subRepo *models.SubscriberRepository,
) *subscriber.Client {
	subConsumer, err := subscriber.NewClient(transportCfg.Config{
		QueueName: userQueueName,
		BrokerURI: config.BrokerURI,
	})
	if err != nil {
		slog.Error("creating subscriber client", slog.Any("error", err))
		return nil
	}
	subConsumer.SubscribeCreate(func(ctx context.Context, email string) error { // nolint: errcheck
		slog.Info("new subscriber", slog.Any("email", email))
		sub := &models.Subscriber{Email: email}
		err := doWithContext(ctx, func() error {
			return subRepo.Create(sub)
		})
		if err != nil {
			slog.Error("saving subscriber", slog.Any("error", err))
		}
		return nil
	})
	subConsumer.SubscribeDelete(func(ctx context.Context, email string) error { // nolint: errcheck
		slog.Info("delete subscriber", slog.Any("email", email))
		subscriber := &models.Subscriber{Email: email}

		err := doWithContext(ctx, func() error {
			return subRepo.Delete(subscriber)
		})
		if err != nil {
			slog.Error("deleting subscriber", slog.Any("error", err))
		}
		return nil
	})
	return subConsumer
}

func main() {
	// Initialize settings
	err := settings.InitSettings()
	if err != nil {
		slog.Error("initializing settings", slog.Any("error", err))
	}

	// Initialize database
	db, err := InitDatabase()
	if err != nil {
		slog.Error("initializing database", slog.Any("error", err))
		panic(err)
	}
	rateRepo := models.NewRateRepository(db)
	subRepo := models.NewSubscriberRepository(db)

	// Initialize mailer by debug mode
	debug := os.Getenv("DEBUG") == "true"
	var mailer mail.Mailer
	mailConfig := mailCfg.NewFromEnv()
	if debug {
		mailer = backends.NewConsoleMailer(mailConfig)
	} else {
		mailer = backends.NewGomailMailer(mailConfig)
	}
	mailClient := mail.NewClient(mailer)

	// Initialize cron manager for notifications
	cronManager := InitNotificationsCron(db, mailClient)
	cronManager.Start()
	defer cronManager.Stop()

	transportConfig := transportCfg.NewFromEnv()
	// Create rate consumer and subscribe to events
	rateConsumer := InitRateConsumer(transportConfig, rateRepo)
	if rateConsumer == nil {
		slog.Error("initializing rate consumer")
	}
	defer rateConsumer.Close()

	// Create subscribe consumer and subscribe to events
	subConsumer := InitSubscriberConsumer(transportConfig, subRepo)
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

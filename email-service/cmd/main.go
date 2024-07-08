package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

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

func InitDatabase() (*database.DB, error) {
	db, err := database.New(dbCfg.NewFromEnv())
	if err != nil {
		return nil, err
	}
	if err := db.Migrate(&models.Rate{}); err != nil {
		return nil, err
	}
	return db, nil
}

func InitNotificationsCron(db *database.DB, mailer notifications.EmailClient) *cron.Manager {
	// Start cron job for notifications
	cronSpec := os.Getenv("CRON_SPEC")
	if cronSpec == "" {
		slog.Warn(
			"CRON_SPEC is not set, using default value",
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

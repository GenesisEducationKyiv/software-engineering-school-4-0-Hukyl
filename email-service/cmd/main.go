package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/pkg/settings"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/broker"
	transportCfg "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/broker/transport/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/mail"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/mail/backends"
	mailCfg "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/mail/config"
)

func main() {
	err := settings.InitSettings()
	if err != nil {
		slog.Error("initializing settings", slog.Any("error", err))
	}

	debug := os.Getenv("DEBUG") == "true"
	var mailer mail.Mailer
	mailConfig := mailCfg.NewFromEnv()
	if debug {
		mailer = backends.NewConsoleMailer(mailConfig)
	} else {
		mailer = backends.NewGomailMailer(mailConfig)
	}
	mailClient := mail.NewClient(mailer)

	transportConfig := transportCfg.NewFromEnv()
	client, err := broker.NewClient(transportConfig)
	if err != nil {
		slog.Error("creating broker client", slog.Any("error", err))
		return
	}
	defer client.Close()

	if err = client.Subscribe(mailClient.SendEmail); err != nil {
		slog.Error("subscribing to broker", slog.Any("error", err))
		return
	}

	termChannel := make(chan os.Signal, 1)
	signal.Notify(termChannel, syscall.SIGINT)
	signal.Notify(termChannel, syscall.SIGTERM)
	<-termChannel
	// Gracefully close the client
}

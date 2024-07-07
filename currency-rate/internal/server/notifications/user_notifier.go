package notifications

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/models"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/server"
)

type EmailClient interface {
	SendEmail(ctx context.Context, emails []string, subject, message string) error
}

type Repository interface {
	FindAll() ([]models.User, error)
}

type RateMessageFormatter interface {
	SetRate(rate *models.Rate)
	Subject() string
	String() string
}

type UsersNotifier struct {
	mailClient       EmailClient
	rateService      server.RateService
	userRepository   Repository
	messageFormatter RateMessageFormatter
}

func NewUsersNotifier(
	mailClient EmailClient,
	rateService server.RateService,
	userRepository Repository,
	msgFormatter RateMessageFormatter,
) *UsersNotifier {
	return &UsersNotifier{
		mailClient:       mailClient,
		rateService:      rateService,
		userRepository:   userRepository,
		messageFormatter: msgFormatter,
	}
}

func logAndWrap(message string, err error) error {
	slog.Error(message, slog.Any("error", err))
	return fmt.Errorf("%s: %w", message, err)
}

func (n *UsersNotifier) Notify(ctx context.Context) error {
	rate, err := n.rateService.FetchRate(ctx, "USD", "UAH")
	if err != nil {
		return logAndWrap("failed to fetch rate", err)
	}

	users, err := n.userRepository.FindAll()
	if err != nil {
		return logAndWrap("failed to fetch users", err)
	}

	slog.Info(
		"notifying users by email",
		slog.Any("userCount", len(users)),
	)

	userEmails := make([]string, 0, len(users))
	for _, user := range users {
		userEmails = append(userEmails, user.Email)
	}

	n.messageFormatter.SetRate(rate)
	err = n.mailClient.SendEmail(
		ctx,
		userEmails,
		n.messageFormatter.Subject(),
		n.messageFormatter.String(),
	)
	if err != nil {
		return logAndWrap("failed to send email", err)
	}
	return nil
}

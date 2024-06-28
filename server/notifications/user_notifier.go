package notifications

import (
	"context"
	"log/slog"

	"github.com/Hukyl/genesis-kma-school-entry/models"
	"github.com/Hukyl/genesis-kma-school-entry/server"
)

type EmailClient interface {
	SendEmail(ctx context.Context, email, subject, message string) error
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

func (n *UsersNotifier) Notify(ctx context.Context) {
	rate, err := n.rateService.FetchRate(ctx, "USD", "UAH")
	if err != nil {
		slog.Warn("failed to fetch rate", slog.Any("error", err))
		return
	}

	users, err := n.userRepository.FindAll()
	if err != nil {
		slog.Error("failed to fetch users", slog.Any("error", err))
		return
	}

	slog.Info(
		"notifying users by email",
		slog.Any("userCount", len(users)),
	)

	n.messageFormatter.SetRate(rate)
	for _, user := range users {
		err := n.mailClient.SendEmail(
			ctx,
			user.Email,
			n.messageFormatter.Subject(),
			n.messageFormatter.String(),
		)
		if err != nil {
			slog.Error(
				"failed sending email",
				slog.Any("error", err),
				slog.Any("userEmail", user.Email),
			)
		}
	}
}

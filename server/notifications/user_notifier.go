package notifications

import (
	"context"
	"log/slog"

	"github.com/Hukyl/genesis-kma-school-entry/models"
	"github.com/Hukyl/genesis-kma-school-entry/server"
)

type EmailClient interface {
	SendEmail(ctx context.Context, email, message string) error
}

type Repository interface {
	FindAll() ([]models.User, error)
}

type UsersNotifier struct {
	mailClient     EmailClient
	rateFetcher    server.RateFetcher
	userRepository Repository
}

func NewUsersNotifier(
	mailClient EmailClient,
	rateFetcher server.RateFetcher,
	userRepository Repository,
) *UsersNotifier {
	return &UsersNotifier{
		mailClient:     mailClient,
		rateFetcher:    rateFetcher,
		userRepository: userRepository,
	}
}

func (n *UsersNotifier) Notify(ctx context.Context) {
	rate, err := n.rateFetcher.FetchRate("USD", "UAH")
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
	for _, user := range users {
		message := NewRateMessage(rate)
		if err := n.mailClient.SendEmail(ctx, user.Email, message.String()); err != nil {
			slog.Error(
				"failed sending email",
				slog.Any("error", err),
				slog.Any("userEmail", user.Email),
			)
		}
	}
}

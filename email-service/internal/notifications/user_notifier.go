package notifications

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/models"
)

type EmailClient interface {
	SendEmail(ctx context.Context, emails []string, subject, message string) error
}

type SubscriberRepository interface {
	FindAll() ([]models.Subscriber, error)
}

type RateRepository interface {
	Latest(ccFrom, ccTo string) (*models.Rate, error)
}

type RateMessageFormatter interface {
	SetRate(rate *models.Rate)
	Subject() string
	String() string
}

type MailNotifier struct {
	mailClient       EmailClient
	rateRepo         RateRepository
	subscriberRepo   SubscriberRepository
	messageFormatter RateMessageFormatter
}

func NewMailNotifier(
	mailClient EmailClient,
	rateRepo RateRepository,
	subscriberRepo SubscriberRepository,
	msgFormatter RateMessageFormatter,
) *MailNotifier {
	return &MailNotifier{
		mailClient:       mailClient,
		rateRepo:         rateRepo,
		subscriberRepo:   subscriberRepo,
		messageFormatter: msgFormatter,
	}
}

func logAndWrap(message string, err error) error {
	slog.Error(message, slog.Any("error", err))
	return fmt.Errorf("%s: %w", message, err)
}

func (n *MailNotifier) Notify(ctx context.Context) error {
	rate, err := n.rateRepo.Latest("USD", "UAH")
	if err != nil {
		return logAndWrap("failed to fetch rate", err)
	}

	subs, err := n.subscriberRepo.FindAll()
	if err != nil {
		return logAndWrap("failed to fetch subscribers", err)
	}

	slog.Info(
		"notifying subscribers by email",
		slog.Any("subCount", len(subs)),
	)

	subEmails := make([]string, 0, len(subs))
	for _, user := range subs {
		subEmails = append(subEmails, user.Email)
	}

	n.messageFormatter.SetRate(rate)
	err = n.mailClient.SendEmail(
		ctx,
		subEmails,
		n.messageFormatter.Subject(),
		n.messageFormatter.String(),
	)
	if err != nil {
		return logAndWrap("failed to send email", err)
	}
	return nil
}

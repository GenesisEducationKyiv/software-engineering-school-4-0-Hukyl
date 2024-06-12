package notifications_test

import (
	"context"
	"testing"

	"github.com/Hukyl/genesis-kma-school-entry/models"
	"github.com/Hukyl/genesis-kma-school-entry/rate"
	"github.com/Hukyl/genesis-kma-school-entry/server/notifications"
	"github.com/stretchr/testify/assert"
)

type mockRateFetcher struct {
	callCount int
}

func (m *mockRateFetcher) FetchRate(from, to string) (rate.Rate, error) {
	m.callCount++
	return rate.Rate{CurrencyFrom: from, CurrencyTo: to, Rate: 27.5}, nil
}

type mockUserRepository struct {
	callCount int
}

func (m *mockUserRepository) FindAll() ([]models.User, error) {
	m.callCount++
	return []models.User{
		{Email: "example@gmail.com"},
		{Email: "example2@gmail.com"},
	}, nil
}

type mockEmailClient struct {
	callCount int
}

func (m *mockEmailClient) SendEmail(_ context.Context, _, _ string) error {
	m.callCount++
	return nil
}

func TestNotify(t *testing.T) {
	ctx := context.Background()
	rateFetcher := &mockRateFetcher{}
	userRepository := &mockUserRepository{}
	emailClient := &mockEmailClient{}
	notifier := notifications.NewUsersNotifier(emailClient, rateFetcher, userRepository)
	notifier.Notify(ctx)
	assert.Equal(t, 1, rateFetcher.callCount)
	assert.Equal(t, 1, userRepository.callCount)
	assert.Equal(t, 2, emailClient.callCount)
}

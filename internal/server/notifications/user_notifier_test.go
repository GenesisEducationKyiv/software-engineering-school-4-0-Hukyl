package notifications_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Hukyl/genesis-kma-school-entry/internal/models"
	"github.com/Hukyl/genesis-kma-school-entry/internal/server/notifications"
	"github.com/stretchr/testify/mock"
)

type mockRateFetcher struct {
	mock.Mock
}

func (m *mockRateFetcher) FetchRate(ctx context.Context, from, to string) (*models.Rate, error) {
	args := m.Called(ctx, from, to)
	return args.Get(0).(*models.Rate), args.Error(1)
}

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) FindAll() ([]models.User, error) {
	args := m.Called()
	return args.Get(0).([]models.User), args.Error(1)
}

type mockMessageFormatter struct {
	mock.Mock
}

func (m *mockMessageFormatter) SetRate(rate *models.Rate) {
	m.Called(rate)
}

func (m *mockMessageFormatter) Subject() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockMessageFormatter) String() string {
	args := m.Called()
	return args.String(0)
}

type mockEmailClient struct {
	mock.Mock
}

func (m *mockEmailClient) SendEmail(ctx context.Context, emails []string, subject, message string) error {
	args := m.Called(ctx, emails, subject, message)
	return args.Error(0)
}

func TestUserNotify(t *testing.T) {
	// Arrange
	ctx := context.Background()
	ccFrom := "USD"
	ccTo := "UAH"

	rateService := new(mockRateFetcher)
	rateService.On("FetchRate", mock.Anything, ccFrom, ccTo).Return(&models.Rate{
		Rate:         27.5,
		CurrencyFrom: ccFrom,
		CurrencyTo:   ccTo,
	}, nil)

	userRepository := new(mockUserRepository)
	userRepository.On("FindAll").Return([]models.User{
		{Email: "example@gmail.com"},
		{Email: "example2@gmail.com"},
	}, nil)

	emailClient := new(mockEmailClient)
	emailClient.On(
		"SendEmail", ctx, []string{"example@gmail.com", "example2@gmail.com"}, mock.Anything, mock.Anything,
	).Return(nil).Once()

	messageFormatter := new(mockMessageFormatter)
	messageFormatter.On("SetRate", &models.Rate{
		Rate:         27.5,
		CurrencyFrom: ccFrom,
		CurrencyTo:   ccTo,
	}).Return()
	messageFormatter.On("Subject").Return(fmt.Sprintf("%s-%s exchange rate", ccFrom, ccTo))
	messageFormatter.On("String").Return(fmt.Sprintf("1 %s = 27.5 %s", ccFrom, ccTo))

	notifier := notifications.NewUsersNotifier(
		emailClient,
		rateService,
		userRepository,
		messageFormatter,
	)
	// Act
	notifier.Notify(ctx)
	// Assert
	rateService.AssertExpectations(t)
	userRepository.AssertExpectations(t)
	emailClient.AssertExpectations(t)
}

package notifications_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/models"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/notifications"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockRateRepo struct {
	mock.Mock
}

func (m *mockRateRepo) Latest(ccFrom, ccTo string) (*models.Rate, error) {
	args := m.Called(ccFrom, ccTo)
	return args.Get(0).(*models.Rate), args.Error(1)
}

type mockSubRepo struct {
	mock.Mock
}

func (m *mockSubRepo) FindAll() ([]models.Subscriber, error) {
	args := m.Called()
	return args.Get(0).([]models.Subscriber), args.Error(1)
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

	rateRepo := new(mockRateRepo)
	rateRepo.On("Latest", ccFrom, ccTo).Return(&models.Rate{
		Rate:         27.5,
		CurrencyFrom: ccFrom,
		CurrencyTo:   ccTo,
	}, nil)

	subRepo := new(mockSubRepo)
	subRepo.On("FindAll").Return([]models.Subscriber{
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

	notifier := notifications.NewMailNotifier(
		emailClient,
		rateRepo,
		subRepo,
		messageFormatter,
	)
	// Act
	err := notifier.Notify(ctx)
	// Assert
	require.NoError(t, err)
	rateRepo.AssertExpectations(t)
	subRepo.AssertExpectations(t)
	emailClient.AssertExpectations(t)
}

package notifications_test

import (
	"context"
	"testing"

	"github.com/Hukyl/genesis-kma-school-entry/models"
	"github.com/Hukyl/genesis-kma-school-entry/server/notifications"
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

func (m *mockEmailClient) SendEmail(ctx context.Context, email, subject, message string) error {
	args := m.Called(ctx, email, subject, message)
	return args.Error(0)
}

func TestUserNotify(t *testing.T) {
	// Arrange
	ctx := context.Background()

	rateService := new(mockRateFetcher)
	rateService.On("FetchRate", mock.Anything, "USD", "UAH").Return(&models.Rate{Rate: 27.5}, nil)

	userRepository := new(mockUserRepository)
	userRepository.On("FindAll").Return([]models.User{
		{Email: "example@gmail.com"},
		{Email: "example2@gmail.com"},
	}, nil)

	emailClient := new(mockEmailClient)
	emailClient.On(
		"SendEmail", ctx, "example@gmail.com", mock.Anything, mock.Anything,
	).Return(nil).Once()
	emailClient.On(
		"SendEmail", ctx, "example2@gmail.com", mock.Anything, mock.Anything,
	).Return(nil).Once()

	messageFormatter := new(mockMessageFormatter)
	messageFormatter.On("SetRate", &models.Rate{Rate: 27.5}).Return()
	messageFormatter.On("Subject").Return("USD-UAH exchange rate")
	messageFormatter.On("String").Return("1 USD = 27.5 UAH")

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

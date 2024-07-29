package mail_test

import (
	"context"
	"errors"
	"testing"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/mail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockBackend struct {
	mock.Mock
}

func (m *mockBackend) SendEmail(ctx context.Context, to []string, subject, message string) error {
	args := m.Called(ctx, to, subject, message)
	return args.Error(0)
}

func TestSendEmail(t *testing.T) {
	testCases := []struct {
		name        string
		err         error
		expectError bool
	}{
		{
			name:        "success",
			err:         nil,
			expectError: false,
		},
		{
			name:        "error",
			err:         errors.New("error"),
			expectError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mb := &mockBackend{}
			recipients := []string{"example@gmail.com", "example2@gmail.com"}
			client := mail.NewClient(mb)
			mb.On("SendEmail", mock.Anything, recipients, "subject", "message").Return(tc.err)
			err := client.SendEmail(context.Background(), recipients, "subject", "message")
			if tc.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

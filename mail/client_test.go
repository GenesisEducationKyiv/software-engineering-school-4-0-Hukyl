package mail_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Hukyl/genesis-kma-school-entry/mail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockBackend struct {
	mock.Mock
}

func (m *mockBackend) SendEmail(ctx context.Context, to, subject, message string) error {
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
			client := mail.NewClient(mb)
			mb.On("SendEmail", mock.Anything, "example@gmail.com", "subject", "message").Return(tc.err)
			err := client.SendEmail(context.Background(), "example@gmail.com", "subject", "message")
			if tc.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

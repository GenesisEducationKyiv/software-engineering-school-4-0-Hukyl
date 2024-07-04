package backends_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/mail"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/mail/backends"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/mail/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendEmail_InvalidPort(t *testing.T) {
	testCases := []struct {
		name        string
		port        string
		expectError bool
	}{
		{
			name:        "invalid port",
			port:        "invalid",
			expectError: true,
		},
		{
			name:        "empty port",
			port:        "",
			expectError: true,
		},
		{
			name:        "no smtp server on port",
			port:        "25",
			expectError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			gm := backends.NewGomailMailer(config.Config{
				FromEmail:    "example@gmail.com",
				SMTPHost:     "smtp.gmail.com",
				SMTPPort:     tc.port,
				SMTPUser:     "user",
				SMTPPassword: "password",
			})
			err := gm.SendEmail(ctx, []string{"example2@gmail.com"}, "subject", "message")
			if tc.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestSendEmail_Success(t *testing.T) {
	testCases := []struct {
		name        string
		toEmails    []string
		expectError bool
	}{
		{
			name:        "single",
			toEmails:    []string{"example2@gmail.com"},
			expectError: false,
		},
		{
			name:        "multiple",
			toEmails:    []string{"example2@gmail.com", "example3@gmail.com", "example4@gmail.com"},
			expectError: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			smtpServer := mail.MockSMTPServer(t)
			ctx := context.Background()
			gm := backends.NewGomailMailer(config.Config{
				FromEmail:    "example@gmail.com",
				SMTPHost:     mail.Localhost,
				SMTPPort:     strconv.Itoa(smtpServer.PortNumber()),
				SMTPUser:     "user",
				SMTPPassword: "password",
			})
			err := gm.SendEmail(ctx, tc.toEmails, "subject", "message")
			if tc.expectError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			messages := smtpServer.Messages()
			assert.Len(t, messages, 1)
		})
	}
}

func TestVariousParameters(t *testing.T) {
	testCases := []struct {
		name        string
		fromEmail   string
		toEmails    []string
		subject     string
		message     string
		expectError bool
	}{
		{
			name:        "valid",
			fromEmail:   "example@gmail.com",
			toEmails:    []string{"example2@gmail.com"},
			subject:     "subject",
			message:     "message",
			expectError: false,
		},
		{
			name:        "valid-multiple",
			fromEmail:   "example@gmail.com",
			toEmails:    []string{"example2@gmail.com", "example3@gmail.com"},
			subject:     "subject",
			message:     "message",
			expectError: false,
		},
		{
			name:        "invalid from email",
			fromEmail:   "example",
			toEmails:    []string{"example@gmail.com"},
			subject:     "subject",
			message:     "message",
			expectError: true,
		},
		{
			name:        "invalid to email",
			fromEmail:   "example@gmail.com",
			toEmails:    []string{"example"},
			subject:     "subject",
			message:     "message",
			expectError: true,
		},
		{
			name:        "empty message",
			fromEmail:   "example@gmail.com",
			toEmails:    []string{"example2@gmail.com"},
			subject:     "",
			message:     "",
			expectError: false,
		},
		{
			name:        "empty subject",
			fromEmail:   "example@gmail.com",
			toEmails:    []string{"example2@gmail.com"},
			subject:     "",
			message:     "message",
			expectError: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			smtpServer := mail.MockSMTPServer(t)
			gm := backends.NewGomailMailer(config.Config{
				FromEmail:    tc.fromEmail,
				SMTPHost:     mail.Localhost,
				SMTPPort:     strconv.Itoa(smtpServer.PortNumber()),
				SMTPUser:     "user",
				SMTPPassword: "password",
			})
			err := gm.SendEmail(ctx, tc.toEmails, tc.subject, tc.message)
			if tc.expectError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, smtpServer.Messages(), 1)
		})
	}
}

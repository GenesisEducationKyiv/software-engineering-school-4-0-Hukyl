package backends_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/Hukyl/genesis-kma-school-entry/mail"
	"github.com/Hukyl/genesis-kma-school-entry/mail/backends"
	"github.com/Hukyl/genesis-kma-school-entry/mail/config"
	"github.com/stretchr/testify/assert"
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
			err := gm.SendEmail(ctx, "example2@gmail.com", "subject", "message")
			if tc.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestSendEmail_Success(t *testing.T) {
	smtpServer := mail.MockSMTPServer(t)
	ctx := context.Background()
	gm := backends.NewGomailMailer(config.Config{
		FromEmail:    "example@gmail.com",
		SMTPHost:     "smtp.gmail.com",
		SMTPPort:     strconv.Itoa(smtpServer.PortNumber()),
		SMTPUser:     "user",
		SMTPPassword: "password",
	})
	err := gm.SendEmail(ctx, "example2@gmail.com", "subject", "message")
	assert.NoError(t, err)
	assert.Len(t, smtpServer.Messages(), 1)
}

func TestVariousParameters(t *testing.T) {
	testCases := []struct {
		name        string
		fromEmail   string
		toEmail     string
		subject     string
		message     string
		expectError bool
	}{
		{
			name:        "valid",
			fromEmail:   "example@gmail.com",
			toEmail:     "example2@gmail.com",
			subject:     "subject",
			message:     "message",
			expectError: false,
		},
		{
			name:        "invalid from email",
			fromEmail:   "example",
			toEmail:     "example@gmail.com",
			subject:     "subject",
			message:     "message",
			expectError: true,
		},
		{
			name:        "invalid to email",
			fromEmail:   "example@gmail.com",
			toEmail:     "example",
			subject:     "subject",
			message:     "message",
			expectError: true,
		},
		{
			name:        "empty message",
			fromEmail:   "example@gmail.com",
			toEmail:     "example2@gmail.com",
			subject:     "",
			message:     "",
			expectError: false,
		},
		{
			name:        "empty subject",
			fromEmail:   "example@gmail.com",
			toEmail:     "example2@gmail.com",
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
			err := gm.SendEmail(ctx, tc.toEmail, tc.subject, tc.message)
			if tc.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, smtpServer.Messages(), 1)
		})
	}
}

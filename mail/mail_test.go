package mail_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Hukyl/genesis-kma-school-entry/mail"
	"github.com/Hukyl/genesis-kma-school-entry/mail/config"
	"github.com/Hukyl/genesis-kma-school-entry/settings"
	"github.com/stretchr/testify/assert"
)

func TestClientSendEmailByDebugKey(t *testing.T) {
	testCases := []struct {
		name            string
		debugKeyPresent bool
		messagesCount   int
	}{
		{
			name:            "with-key",
			debugKeyPresent: true,
			messagesCount:   0,
		},
		{
			name:            "without-key",
			debugKeyPresent: false,
			messagesCount:   1,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			if tc.debugKeyPresent {
				ctx = context.WithValue(ctx, settings.DebugKey, true)
			}
			smtpServer := mail.MockSMTPServer(t)
			mc := mail.Client{
				Config: config.Config{
					FromEmail:    "example@gmail.com",
					SMTPHost:     mail.Localhost,
					SMTPPort:     fmt.Sprint(smtpServer.PortNumber()),
					SMTPUser:     "user",
					SMTPPassword: "password",
				},
			}
			err := mc.SendEmail(ctx, "example2@gmail.com", "subject", "message")
			assert.NoError(t, err)
			assert.Len(t, smtpServer.Messages(), tc.messagesCount)
		})
	}
}

func TestClientSendEmailStub(t *testing.T) {
	mc := mail.Client{
		Config: config.Config{
			FromEmail: "example@gmail.com",
		},
	}
	err := mc.SendEmailStub("example2@gmail.com", "subject", "message")
	assert.NoError(t, err)
}

func TestClientSMTPEmailInvalidPort(t *testing.T) {
	mc := mail.Client{
		Config: config.Config{
			FromEmail:    "example@gmail.com",
			SMTPHost:     "smtp.gmail.com",
			SMTPPort:     "invalid",
			SMTPUser:     "user",
			SMTPPassword: "password",
		},
	}
	err := mc.SendSMTPEmail("example2@gmail.com", "subject", "message")
	assert.Error(t, err)
}

func TestClientSMTPEmail(t *testing.T) {
	smtpServer := mail.MockSMTPServer(t)
	mc := mail.Client{
		Config: config.Config{
			FromEmail:    "example@gmail.com",
			SMTPHost:     mail.Localhost,
			SMTPPort:     fmt.Sprint(smtpServer.PortNumber()),
			SMTPUser:     "user",
			SMTPPassword: "password",
		},
	}
	err := mc.SendSMTPEmail("example2@gmail.com", "subject", "message")
	assert.NoError(t, err)
	assert.Len(t, smtpServer.Messages(), 1)
}

func TestClientSMTPEmailVariousParameters(t *testing.T) {
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
			smtpServer := mail.MockSMTPServer(t)
			mc := mail.Client{
				Config: config.Config{
					FromEmail:    tc.fromEmail,
					SMTPHost:     mail.Localhost,
					SMTPPort:     fmt.Sprint(smtpServer.PortNumber()),
					SMTPUser:     "user",
					SMTPPassword: "password",
				},
			}
			err := mc.SendSMTPEmail(tc.toEmail, tc.subject, tc.message)
			if tc.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, smtpServer.Messages(), 1)
		})
	}
}

func TestClientSMTPEmailTimeout(t *testing.T) {
	smtpServer := mail.MockSMTPServer(t)
	mc := mail.Client{
		Config: config.Config{
			FromEmail:    "example@gmail.com",
			SMTPHost:     mail.Localhost,
			SMTPPort:     fmt.Sprint(smtpServer.PortNumber() + 1),
			SMTPUser:     "user",
			SMTPPassword: "password",
		},
	}
	err := mc.SendSMTPEmail("example2@gmail.com", "subject", "message")
	assert.Error(t, err)
	assert.Len(t, smtpServer.Messages(), 0)
}

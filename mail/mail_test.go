package mail_test

import (
	"fmt"
	"testing"

	"github.com/Hukyl/genesis-kma-school-entry/mail"
	"github.com/Hukyl/genesis-kma-school-entry/mail/config"
	smtpmock "github.com/mocktools/go-smtp-mock/v2"
	"github.com/stretchr/testify/assert"
)

const localhost = "127.0.0.1"

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
	smtpServer := smtpmock.New(smtpmock.ConfigurationAttr{})
	if err := smtpServer.Start(); err != nil {
		t.Error("failed to start smtp server")
	}
	defer func() {
		if err := smtpServer.Stop(); err != nil {
			t.Error("failed to stop smtp server")
		}
	}()
	hostAddress, portNumber := localhost, smtpServer.PortNumber()

	mc := mail.Client{
		Config: config.Config{
			FromEmail:    "example@gmail.com",
			SMTPHost:     hostAddress,
			SMTPPort:     fmt.Sprint(portNumber),
			SMTPUser:     "user",
			SMTPPassword: "password",
		},
	}
	err := mc.SendSMTPEmail("example2@gmail.com", "subject", "message")
	assert.NoError(t, err)
	assert.Len(t, smtpServer.Messages(), 1)
}

func TestClientSMTPEmailVariousParameters(t *testing.T) { // nolint: funlen
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
			smtpServer := smtpmock.New(smtpmock.ConfigurationAttr{})
			if err := smtpServer.Start(); err != nil {
				t.Error("failed to start smtp server")
			}
			defer func() {
				if err := smtpServer.Stop(); err != nil {
					t.Error("failed to stop smtp server")
				}
			}()
			hostAddress, portNumber := localhost, smtpServer.PortNumber()

			mc := mail.Client{
				Config: config.Config{
					FromEmail:    tc.fromEmail,
					SMTPHost:     hostAddress,
					SMTPPort:     fmt.Sprint(portNumber),
					SMTPUser:     "user",
					SMTPPassword: "password",
				},
			}
			err := mc.SendSMTPEmail(tc.toEmail, tc.subject, tc.message)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, smtpServer.Messages(), 1)
			}
		})
	}
}

func TestClientSMTPEmailTimeout(t *testing.T) {
	smtpServer := smtpmock.New(smtpmock.ConfigurationAttr{})
	if err := smtpServer.Start(); err != nil {
		t.Error("failed to start smtp server")
	}
	defer func() {
		if err := smtpServer.Stop(); err != nil {
			t.Error("failed to stop smtp server")
		}
	}()
	hostAddress, portNumber := localhost, smtpServer.PortNumber()

	mc := mail.Client{
		Config: config.Config{
			FromEmail:    "example@gmail.com",
			SMTPHost:     hostAddress,
			SMTPPort:     fmt.Sprint(portNumber + 1),
			SMTPUser:     "user",
			SMTPPassword: "password",
		},
	}
	err := mc.SendSMTPEmail("example2@gmail.com", "subject", "message")
	assert.Error(t, err)
	assert.Len(t, smtpServer.Messages(), 0)
}

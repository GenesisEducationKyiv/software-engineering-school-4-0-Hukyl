package backends_test

import (
	"context"
	"testing"

	"github.com/Hukyl/genesis-kma-school-entry/internal/mail/backends"
	"github.com/Hukyl/genesis-kma-school-entry/internal/mail/config"
	"github.com/stretchr/testify/assert"
)

func TestClientSendEmailStub(t *testing.T) {
	c := backends.NewConsoleMailer(config.Config{
		FromEmail: "example@gmail.com",
	})
	err := c.SendEmail(context.Background(), "example2@gmail.com", "subject", "message")
	assert.NoError(t, err)
}

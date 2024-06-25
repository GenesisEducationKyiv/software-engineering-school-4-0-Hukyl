package mail

import (
	"testing"

	smtpmock "github.com/mocktools/go-smtp-mock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const Localhost = "127.0.0.1"

func MockSMTPServer(t *testing.T) *smtpmock.Server {
	t.Helper()
	smtpServer := smtpmock.New(smtpmock.ConfigurationAttr{
		HostAddress: Localhost,
	})
	err := smtpServer.Start()
	require.NoError(t, err, "failed to start smtp server")
	t.Cleanup(func() {
		err := smtpServer.Stop()
		assert.NoError(t, err, "failed to stop smtp server")
	})
	return smtpServer
}

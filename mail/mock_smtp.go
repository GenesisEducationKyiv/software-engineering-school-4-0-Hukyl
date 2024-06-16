package mail

import (
	"testing"

	smtpmock "github.com/mocktools/go-smtp-mock/v2"
)

const Localhost = "127.0.0.1"

func MockSMTPServer(t *testing.T) *smtpmock.Server {
	t.Helper()
	smtpServer := smtpmock.New(smtpmock.ConfigurationAttr{
		HostAddress: Localhost,
	})
	if err := smtpServer.Start(); err != nil {
		t.Error("failed to start smtp server")
	}
	t.Cleanup(func() {
		if err := smtpServer.Stop(); err != nil {
			t.Error("failed to stop smtp server")
		}
	})
	return smtpServer
}

package cron

import (
	"fmt"
	"log/slog"

	"github.com/robfig/cron/v3"
)

var logger *slog.Logger

func getLogger() *slog.Logger {
	if logger == nil {
		logger = slog.Default().With(slog.Any("src", "cronManager"))
	}
	return logger
}

type Doer func() error

type Manager struct {
	cron *cron.Cron
}

func NewManager() *Manager {
	return &Manager{
		cron: cron.New(),
	}
}

func (m *Manager) intercept(f Doer) func() {
	return func() {
		if err := f(); err != nil {
			getLogger().Error("cron job failed", slog.Any("error", err), slog.Any("job", f))
			return
		}
		getLogger().Debug("cron job done successfully", slog.Any("job", f))
	}
}

func (m *Manager) AddJob(spec string, f Doer) error {
	_, err := m.cron.AddFunc(spec, m.intercept(f))
	if err != nil {
		getLogger().Error("failed to add cron job", slog.Any("error", err))
		return fmt.Errorf("failed to add cron job: %w", err)
	}
	getLogger().Info("cron job added", slog.Any("spec", spec))
	return nil
}

func (m *Manager) Start() {
	getLogger().Info("cron manager started")
	m.cron.Start()
}

func (m *Manager) Stop() {
	getLogger().Info("cron manager stopped")
	m.cron.Stop()
}

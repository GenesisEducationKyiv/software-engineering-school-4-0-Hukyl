package cron

import (
	"fmt"
	"log/slog"

	"github.com/robfig/cron/v3"
)

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
			slog.Error("cron job failed", slog.Any("error", err))
		}
	}
}

func (m *Manager) AddJob(spec string, f Doer) error {
	_, err := m.cron.AddFunc(spec, m.intercept(f))
	if err != nil {
		slog.Error("failed to add cron job", slog.Any("error", err))
		return fmt.Errorf("failed to add cron job: %w", err)
	}
	slog.Info("cron job added", slog.Any("spec", spec))
	return nil
}

func (m *Manager) Start() {
	m.cron.Start()
}

func (m *Manager) Stop() {
	m.cron.Stop()
}

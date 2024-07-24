package main

import (
	"log/slog"
	"os"
)

func NewLogger() *slog.Logger {
	loggerOptions := &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}
	logger := slog.New(
		slog.NewTextHandler(os.Stdout, loggerOptions),
	).With(slog.Any("service", "currency-rate"))
	return logger
}

func main() {
	// logger := NewLogger()
	slog.SetDefault(NewLogger())

	defaultLogger := slog.Default().With(slog.Any("src", "main"))
	defaultLogger.Debug("Hello, World!", slog.Any("key", "value"))
}

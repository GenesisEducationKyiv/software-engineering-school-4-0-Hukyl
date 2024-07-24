package settings

import (
	"fmt"
	"log/slog"

	"github.com/joho/godotenv"
)

func InitSettings(path ...string) error {
	var logger = slog.Default().With(slog.Any("src", "settings"))

	if len(path) == 0 {
		path = append(path, ".env")
	}
	logger.Debug("initializing settings", slog.Any("path", path))
	for _, p := range path {
		err := godotenv.Load(p)
		if err != nil {
			return fmt.Errorf("init settings: %w", err)
		}
	}
	return nil
}

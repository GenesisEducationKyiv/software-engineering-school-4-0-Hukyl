package settings

import (
	"fmt"

	"github.com/joho/godotenv"
)

func InitSettings() error {
	err := godotenv.Load(".env")
	if err != nil {
		return fmt.Errorf("init settings: %w", err)
	}
	return nil
}

type ContextKey int

const (
	DebugKey ContextKey = iota
)

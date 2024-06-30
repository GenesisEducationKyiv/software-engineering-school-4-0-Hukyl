package settings

import (
	"fmt"

	"github.com/joho/godotenv"
)

func InitSettings(path ...string) error {
	if len(path) == 0 {
		path = append(path, ".env")
	}
	for _, p := range path {
		err := godotenv.Load(p)
		if err != nil {
			return fmt.Errorf("init settings: %w", err)
		}
	}
	return nil
}

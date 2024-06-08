package utils

import (
	"github.com/Hukyl/genesis-kma-school-entry/models"
	"github.com/Hukyl/genesis-kma-school-entry/settings"
)

func SetUpTestDB() {
	settings.Debug = true
	// NOTE: only works with SQLite in-memory databases
	settings.DatabaseService = "sqlite"
	settings.DatabaseDSN = "file::memory:?cache=shared"
	models.NewDB().Migrate()
}

func TearDownTestDB() {
	models.NewDB().Connection().Migrator().DropTable(&models.User{})
}

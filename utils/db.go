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
	if err := models.NewDB().Migrate(); err != nil {
		// test should fail if this happens
		panic(err)
	}
}

func TearDownTestDB() {
	err := models.NewDB().Connection().Migrator().DropTable(&models.User{})
	if err != nil {
		// test should fail if this happens
		panic(err)
	}
}

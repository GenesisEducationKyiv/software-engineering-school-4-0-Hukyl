package models

import (
	"testing"

	"github.com/Hukyl/genesis-kma-school-entry/models/config"
)

func SetUpTestDB(tb testing.TB) *DB {
	tb.Helper()
	config := config.Config{
		DatabaseService: "sqlite",
		DatabaseDSN:     "file::memory:?cache=shared",
	}
	db := NewDB(config)
	// NOTE: only works with SQLite in-memory databases
	if err := db.Migrate(); err != nil {
		tb.Errorf("Failed to migrate database: %v", err)
	}
	tb.Cleanup(func() {
		migrator := db.Connection().Migrator()
		if err := migrator.DropTable(&User{}); err != nil {
			tb.Errorf("Failed to drop table: %v", err)
		}
	})
	return db
}

package database

import (
	"testing"

	"github.com/Hukyl/genesis-kma-school-entry/database/config"
)

func SetUpTest(tb testing.TB, models ...any) *DB {
	tb.Helper()
	config := config.Config{
		DatabaseService: "sqlite",
		DatabaseDSN:     "file::memory:?cache=shared",
	}
	db, err := New(config)
	if err != nil {
		tb.Fatalf("failed to create database: %v", err)
	}
	// NOTE: only works with SQLite in-memory databases
	if err := db.Migrate(models...); err != nil {
		tb.Errorf("failed to migrate database: %v", err)
	}
	tb.Cleanup(func() {
		migrator := db.Connection().Migrator()
		for _, model := range models {
			if err := migrator.DropTable(model); err != nil {
				tb.Errorf("failed to drop table: %v", err)
			}
		}
	})
	return db
}

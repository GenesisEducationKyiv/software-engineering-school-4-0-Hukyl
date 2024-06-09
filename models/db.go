package models

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/Hukyl/genesis-kma-school-entry/models/config"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DB struct {
	Config config.Config
	conn   *gorm.DB
}

func openConnection(service, dsn string) (gorm.Dialector, error) {
	var open func(string) gorm.Dialector
	switch service {
	case "sqlite":
		open = sqlite.Open
	case "postgres":
		open = postgres.Open
	default:
		return nil, errors.New("unknown database service")
	}
	return open(dsn), nil
}

func (d *DB) Connection() *gorm.DB {
	config := d.Config
	if d.conn == nil {
		dialect, err := openConnection(
			config.DatabaseService,
			config.DatabaseDSN,
		)
		if err != nil {
			slog.Error(fmt.Sprintf(
				"Unknown database service: %s",
				config.DatabaseService,
			))
			return nil
		} else if dialect == nil {
			slog.Error(fmt.Sprintf(
				"Failed to open connection to database: %s",
				config.DatabaseService,
			))
			return nil
		}
		db, err := gorm.Open(dialect, &gorm.Config{})
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to connect to database: %s", err))
			return nil
		}
		d.conn = db
		slog.Info(fmt.Sprintf(
			"Opening connection to db (%s, %s)",
			config.DatabaseService, config.DatabaseDSN,
		))
	}
	return d.conn
}

func (d *DB) Init() {
	// Initialize first connection with the database
	d.Connection()
}

func (d *DB) Migrate() error {
	db := d.Connection()
	err := db.AutoMigrate(&User{})
	if err != nil {
		return fmt.Errorf("Failed to migrate User: %w", err)
	}
	return nil
}

func NewDB(c config.Config) *DB {
	db := DB{Config: c}
	db.Init()
	return &db
}

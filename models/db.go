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
	if d.conn != nil {
		return d.conn
	}
	dialect, err := openConnection(
		config.DatabaseService,
		config.DatabaseDSN,
	)
	if err != nil {
		slog.Error(
			"unknown database service",
			slog.Any("databaseService", config.DatabaseService),
		)
		return nil
	} else if dialect == nil {
		slog.Error(
			"failed to open connection to database",
			slog.Any("databaseService", config.DatabaseService),
		)
		return nil
	}
	db, err := gorm.Open(dialect, &gorm.Config{})
	if err != nil {
		slog.Error("failed to connect to database", slog.Any("error", err))
		return nil
	}
	d.conn = db
	slog.Info(
		"opening connection to db",
		slog.Any("databaseService", config.DatabaseService),
		slog.Any("databaseDSN", config.DatabaseDSN),
	)
	return d.conn
}

func (d *DB) Init() error {
	// Initialize first connection with the database
	conn := d.Connection()
	if conn == nil {
		return fmt.Errorf("failed to connect to database")
	}
	return nil
}

func (d *DB) Migrate() error {
	db := d.Connection()
	err := db.AutoMigrate(&User{})
	if err != nil {
		slog.Error("failed to migrate User", slog.Any("error", err))
		return fmt.Errorf("failed to migrate User: %w", err)
	}
	return nil
}

func NewDB(c config.Config) (*DB, error) {
	db := DB{Config: c}
	err := db.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize db: %w", err)
	}
	return &db, nil
}

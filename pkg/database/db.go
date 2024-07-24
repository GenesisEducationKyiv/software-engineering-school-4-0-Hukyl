package database

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/database/config"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var logger *slog.Logger

func getLogger() *slog.Logger {
	if logger == nil {
		logger = slog.Default().With(slog.Any("src", "database"))
	}
	return logger
}

type DB struct {
	Config config.Config
	conn   *gorm.DB
}

func openConnection(service, dsn string) (gorm.Dialector, error) {
	var open func(string) gorm.Dialector
	getLogger().Debug("opening connection to db", slog.Any("databaseService", service))
	switch service {
	case "sqlite":
		open = sqlite.Open
	case "postgres":
		open = postgres.Open
	default:
		return nil, errors.New("unknown database service")
	}
	getLogger().Debug("connection opened")
	return open(dsn), nil
}

func (d *DB) Connection() *gorm.DB {
	config := d.Config
	if d.conn != nil {
		getLogger().Debug("returning existing connection")
		return d.conn
	}
	dialect, err := openConnection(
		config.DatabaseService,
		config.DatabaseDSN,
	)
	if err != nil {
		getLogger().Error(
			"unknown database service",
			slog.Any("databaseService", config.DatabaseService),
		)
		return nil
	} else if dialect == nil {
		getLogger().Error(
			"failed to open connection to database",
			slog.Any("databaseService", config.DatabaseService),
		)
		return nil
	}
	getLogger().Debug("connection opened successfully")
	conn, err := gorm.Open(dialect, &gorm.Config{})
	if err != nil {
		getLogger().Error("failed to connect to database", slog.Any("error", err))
		return nil
	}
	getLogger().Debug("gorm connection established")
	d.conn = conn
	getLogger().Info(
		"opening connection to db",
		slog.Any("databaseService", config.DatabaseService),
		slog.Any("databaseDSN", config.DatabaseDSN),
	)
	return d.conn
}

func (d *DB) Init() error {
	// Initialize first connection with the database
	getLogger().Debug("initializing connection")
	conn := d.Connection()
	if conn == nil {
		getLogger().Error(
			"opening connection to db",
			slog.Any("error", "failed to connect to database"),
		)
		return errors.New("failed to connect to database")
	}
	getLogger().Debug("connection initialized")
	return nil
}

func (d *DB) Migrate(models ...any) error {
	conn := d.Connection()
	if conn == nil {
		return errors.New("failed to connect to database")
	}
	for _, m := range models {
		err := conn.AutoMigrate(m)
		if err != nil {
			getLogger().Error(
				"failed to migrate",
				slog.Any("error", err),
				slog.Any("model", m),
			)
			return fmt.Errorf("failed to migrate %s: %w", m, err)
		}
		getLogger().Debug("migrated", slog.Any("model", m))
	}
	return nil
}

func (d *DB) Close() error {
	if d.conn == nil {
		return nil
	}
	getLogger().Info("closing connection to db")
	sqlDB, err := d.conn.DB()
	if err != nil {
		getLogger().Error("closing database connection", slog.Any("error", err))
		return fmt.Errorf("get db connection error: %w", err)
	}
	if err = sqlDB.Close(); err != nil {
		getLogger().Error("closing database connection", slog.Any("error", err))
		return fmt.Errorf("failed to close connection to database: %w", err)
	}
	return nil
}

func New(c config.Config) (*DB, error) {
	db := DB{Config: c}
	err := db.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize db: %w", err)
	}
	return &db, nil
}

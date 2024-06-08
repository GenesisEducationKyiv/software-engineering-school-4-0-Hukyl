package models

import (
	"errors"
	"log"

	"github.com/Hukyl/genesis-kma-school-entry/settings"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Global singleton database connection
var conn *gorm.DB

type DB interface {
	Connection() *gorm.DB
	Init()
	Migrate() error
}

type db struct{}

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

func (d *db) Connection() *gorm.DB {
	if conn == nil {
		dialect, err := openConnection(settings.DatabaseService, settings.DatabaseDSN)
		if err != nil {
			log.Fatalf("Unknown database service: %s", settings.DatabaseService)
			return nil
		} else if dialect == nil {
			log.Fatalf("Failed to open connection todatabase: %s", settings.DatabaseDSN)
			return nil
		}
		db, err := gorm.Open(dialect, &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to database: %s", err)
			return nil
		}
		conn = db
		log.Printf(
			"Opening connection to db (%s, %s)",
			settings.DatabaseService, settings.DatabaseDSN,
		)
	}
	return conn
}

func (d *db) Init() {
	// Initialize first connection with the database
	d.Connection()
}

func (d *db) Migrate() error {
	db := d.Connection()
	return db.AutoMigrate(&User{})
}

func NewDB() DB {
	db := &db{}
	db.Init()
	return db
}

package settings

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

func InitSettings() {
	godotenv.Load(".env")

	Debug = os.Getenv("DEBUG") == "true"
	if Debug {
		EMAIL_INTERVAL = 1 * time.Minute
	}

	// Webserver configuration
	Port = os.Getenv("PORT")

	// Database
	DatabaseService = os.Getenv("DATABASE_SERVICE")
	DatabaseDSN = os.Getenv("DATABASE_DSN")

	// Email
	SMTPHost = os.Getenv("SMTP_HOST")
	SMTPPort = os.Getenv("SMTP_PORT")
	SMTPUser = os.Getenv("SMTP_USER")
	SMTPPassword = os.Getenv("SMTP_PASSWORD")

	FromEmail = os.Getenv("FROM_EMAIL")
}

var EMAIL_INTERVAL time.Duration = 24 * time.Hour

var Debug bool = false

// Webserver configuration
var Port string = "8080"

// Database
var DatabaseService string = "sqlite"
var DatabaseDSN string = "file::memory:?cache=shared"

// Email
var SMTPHost string = "smtp.gmail.com"
var SMTPPort string = "587"
var SMTPUser string
var SMTPPassword string
var FromEmail string

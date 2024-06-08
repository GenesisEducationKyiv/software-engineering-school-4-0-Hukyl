package utils

import (
	"log"
	"strconv"

	"github.com/Hukyl/genesis-kma-school-entry/settings"
	"github.com/go-gomail/gomail"
)

func SendEmail(email, message string) error {
	if settings.Debug {
		return SendEmailStub(email, message)
	} else {
		return SendSMTPEmail(email, message)
	}
}

func SendEmailStub(email, message string) error {
	log.Printf("Sending email to %s: %s\n", email, message)
	return nil
}

func SendSMTPEmail(email, message string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", settings.FromEmail)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "USD-UAH exchange rate")
	m.SetBody("text/html", message)

	port, _ := strconv.Atoi(settings.SMTPPort)
	d := gomail.NewDialer(
		settings.SMTPHost,
		port,
		settings.SMTPUser,
		settings.SMTPPassword,
	)
	return d.DialAndSend(m)
}

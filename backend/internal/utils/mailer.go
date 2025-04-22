package utils

import (
	"os"

	"gopkg.in/gomail.v2"
)

func SendMail(to string, subject string, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("EMAIL_FROM"))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(
		"smtp.gmail.com",
		587,
		os.Getenv("EMAIL_FROM"),
		os.Getenv("EMAIL_PASSWORD"),
	)

	return d.DialAndSend(m)
}

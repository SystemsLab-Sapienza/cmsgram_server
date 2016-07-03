package main

import (
	"bytes"
	"net/smtp"
	"text/template"
)

func sendEmail(to, subject, body string) (err error) {
	var (
		b    bytes.Buffer
		data = struct {
			Recipient string
			Subject   string
			Body      string
		}{to, subject, body}
	)

	t, err := template.ParseFiles("templates/email.tpl")
	if err != nil {
		return
	}

	err = t.Execute(&b, data)
	if err != nil {
		return
	}

	smtpAuth := smtp.PlainAuth("", Config.EmailUsername, Config.EmailPassword, "smtp.gmail.com")
	err = smtp.SendMail(Config.EmailServer, smtpAuth, Config.EmailUsername, []string{to}, b.Bytes())
	if err != nil {
		return
	}

	return nil
}

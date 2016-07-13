package main

import (
	"bytes"
	"net/smtp"
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

	if config.EmailTestAddress != "" {
		data.Recipient = config.EmailTestAddress
	}

	err = templates.ExecuteTemplate(&b, "email.tpl", data)
	if err != nil {
		return
	}

	smtpAuth := smtp.PlainAuth("", config.EmailUsername, config.EmailPassword, "smtp.gmail.com")
	err = smtp.SendMail(config.EmailServer, smtpAuth, config.EmailUsername, []string{data.Recipient}, b.Bytes())
	if err != nil {
		return
	}

	return nil
}

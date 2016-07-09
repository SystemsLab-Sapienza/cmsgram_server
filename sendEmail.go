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

	if Config.EmailTestAddress != "" {
		data.Recipient = Config.EmailTestAddress
	}

	err = templates.ExecuteTemplate(&b, "email.tpl", data)
	if err != nil {
		return
	}

	smtpAuth := smtp.PlainAuth("", Config.EmailUsername, Config.EmailPassword, "smtp.gmail.com")
	err = smtp.SendMail(Config.EmailServer, smtpAuth, Config.EmailUsername, []string{data.Recipient}, b.Bytes())
	if err != nil {
		return
	}

	return nil
}

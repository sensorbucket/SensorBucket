package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
)

type emailSender struct {
	username string
	password string
	host     string
}

func (infra *emailSender) SendMail(subject string, from string, to string, templateHtml string, content interface{}) error {
	log.Printf("sending mail to '%s'\n", to)
	t, err := template.ParseFiles(templateHtml)
	if err != nil {
		return fmt.Errorf("template parse: %w", err)
	}
	var body bytes.Buffer

	// Add all required headers
	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body.Write([]byte(fmt.Sprintf(
		"From: %s \nSubject: %s \n%s\n\n", from, subject, mimeHeaders)))
	if err = t.Execute(&body, content); err != nil {
		return fmt.Errorf("template execute: %w", err)
	}
	return smtp.SendMail(infra.host, smtp.CRAMMD5Auth(infra.username, infra.password), from, []string{to}, body.Bytes())
}

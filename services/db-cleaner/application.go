package main

import (
	"log"
	"time"
)

type mailer interface {
	SendMail(subject string, from string, to string, templateHtml string, content interface{}) error
}

type store interface {
	DeleteExpiredData() error
}

type service struct {
	mailer mailer
	store  store
}

// Will cleanup any data that have an expired expiration date
func (s *service) clean() error {
	log.Println("running cleanup...")
	return s.store.DeleteExpiredData()
}

// Will send a warning email to configured recipients that their data is scheduled for deletion
func (s *service) warn(from string, recipient string, daysTillDeletion int) error {
	log.Println("warning recipients of scheduled data cleanup...")
	return s.mailer.SendMail(
		"SensorBucket - Geplande gegevensopruiming ",
		from,
		recipient,
		"email-templates/email_template_inlined.html",
		struct {
			DaysTillDeletion int
			DeletionDate     string
		}{
			DeletionDate: time.Now().Add(time.Hour * 24 * time.Duration(daysTillDeletion)).Format("2 January 2006"),
			// TODO: ensure time is displayed in local time to user in html template once emails are a requirement
			DaysTillDeletion: daysTillDeletion,
		})
}

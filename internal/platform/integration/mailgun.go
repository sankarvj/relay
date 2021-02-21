package integration

import (
	"log"

	"github.com/mailgun/mailgun-go"
)

func sendViaMailGun(domain, key, from string, to []string, subject, body string) (*string, error) {
	mg := mailgun.NewMailgun(domain, key)
	m := mg.NewMessage(
		from,
		subject,
		body,
		to...,
	)
	resMsg, id, err := mg.Send(m)
	log.Println("resMsg ", resMsg)
	log.Println("resMsg id ", id)
	log.Println("resMsg err ", err)
	return &id, err
}

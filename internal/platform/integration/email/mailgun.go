package email

import (
	"log"

	"github.com/mailgun/mailgun-go"
)

type MailGun struct {
	Domain    string
	TokenJson string
}

func (m MailGun) SendMail(fromName, fromEmail string, toName string, toEmail []string, subject string, body string) (*string, error) {
	mg := mailgun.NewMailgun(m.Domain, m.TokenJson)
	msg := mg.NewMessage(
		fromEmail,
		subject,
		body,
		toEmail...,
	)
	resMsg, id, err := mg.Send(msg)
	log.Printf("internal.platform.integration.email.mailgun response - resMsg:%s  id: %s err:%v\n", resMsg, id, err)
	return &id, err
}

func (m *MailGun) Watch(topicName string) (string, error) {
	return "", nil
}

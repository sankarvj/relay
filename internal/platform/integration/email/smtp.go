package email

import (
	"crypto/tls"
	"log"

	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	gomail "gopkg.in/mail.v2"
)

type SMTPMail struct {
	Domain  string
	ReplyTo string
}

func (m SMTPMail) SendMail(fromName, fromEmail string, toName string, toEmails []string, subject string, body string) (*string, error) {
	log.Printf("internal.platform.integration.email.fallback request - domain:%s  from: %s\n", m.Domain, fromEmail)
	resMsg, err := send(fromEmail, util.ConvertStrToPtStr(toEmails), subject, body, m.ReplyTo)
	log.Printf("internal.platform.integration.email.fallback response - resMsg:%s  err:%v\n", resMsg, err)
	return resMsg.MessageId, err
}

func (m SMTPMail) Watch(topicName string) (string, error) {
	return "", nil
}

func (m SMTPMail) Stop(emailAddress string) error {
	return nil
}

func sendSMTPMail() error {
	m := gomail.NewMessage()

	// Set E-Mail sender
	m.SetHeader("From", "from@gmail.com")

	// Set E-Mail receivers
	m.SetHeader("To", "to@example.com")

	// Set E-Mail subject
	m.SetHeader("Subject", "Gomail test subject")

	// Set E-Mail body. You can set plain text or html with text/html
	m.SetBody("text/plain", "This is Gomail test body")

	// Settings for SMTP server
	d := gomail.NewDialer("smtp.gmail.com", 587, "from@gmail.com", "<email_password>")

	// This is only needed when SSL/TLS certificate is not valid on server.
	// In production this should be set to false.
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// Now send E-Mail
	if err := d.DialAndSend(m); err != nil {
		log.Println("***> unexpected error occurred in integration.email.smtp. when sending email", err)
		return err
	}
	return nil
}

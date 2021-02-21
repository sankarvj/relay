package integration

import (
	"strings"

	"github.com/pkg/errors"
)

//Integration Types
const (
	TypeGmail = "gmail"
)

//Integration Mail Domains
const (
	DomainMailGun = "mailgun.org"
	DomainGMail   = "google.com"
)

func SendEmail(domain, apiKey, from string, to []string, subject, body string) (*string, error) {
	switch {
	case strings.HasSuffix(domain, DomainMailGun):
		return sendViaMailGun(domain, apiKey, from, to, subject, body)
	case strings.HasSuffix(domain, DomainGMail):
		return sendViaGmail("config/dev/google-apps-client-secret.json", apiKey, "me", "", from, "", to, subject, body)
	default:
		return nil, errors.New("No e-mail client configured to send the mail template")
	}

}

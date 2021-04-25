package integration

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/gmail/v1"
)

//Integration Types
const (
	TypeGmail          = "gmail"
	TypeGoogleCalendar = "google_calendar"
)

//Integration Mail Domains
const (
	DomainMailGun = "mailgun.org"
	DomainGMail   = "google.com"
)

var (
	GmailScopes          = []string{gmail.GmailReadonlyScope, gmail.GmailSendScope}
	GoogleCalendarScopes = []string{calendar.CalendarScope}
)

//GetGoogleAccessURL gets the access-url for the scopes mentioned. This url should be loaded in the UI
func GetGoogleAccessURL(ctx context.Context, oAuthFile string, integId string, scope ...string) (string, error) {
	config, err := getConfig(oAuthFile, scope...)
	if err != nil {
		return "", err
	}
	return config.AuthCodeURL(integId, oauth2.AccessTypeOffline), nil
}

//GetGoogleToken retrives the token for the given scopes. This token should be stored and used for further google API calls
func GetGoogleToken(oAuthFile, code string, scope ...string) (string, error) {
	config, err := getConfig(oAuthFile, scope...)
	if err != nil {
		return "", err
	}
	return tokenJson(config, code)
}

func tokenJson(config *oauth2.Config, code string) (string, error) {
	tok, err := config.Exchange(context.TODO(), code)
	if err != nil {
		return "", errors.Wrap(err, "unable to retrieve token from web")
	}
	tokenJson, err := json.Marshal(tok)
	if err != nil {
		return "", errors.Wrap(err, "unable to marshal token")
	}

	return string(tokenJson), nil
}

func getConfig(oAuthFile string, scope ...string) (*oauth2.Config, error) {
	b, err := ioutil.ReadFile(oAuthFile)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read client secret file")
	}
	return google.ConfigFromJSON(b, scope...)
}

func client(config *oauth2.Config, tokenJson string) (*http.Client, error) {
	var token oauth2.Token
	if err := json.Unmarshal([]byte(tokenJson), &token); err != nil {
		return nil, err
	}
	return config.Client(context.Background(), &token), nil
}

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

type ActionPayload struct {
	ID      string            `json:"id"`
	Payload map[string]string `json:"payload"`
}

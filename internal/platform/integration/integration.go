package integration

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

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

//Google Scopes
var (
	GmailScopes          = []string{gmail.GmailReadonlyScope, gmail.GmailSendScope}
	GoogleCalendarScopes = []string{calendar.CalendarScope}
)

//GetGoogleAccessURL gets the access-url for the scopes mentioned. This url should be loaded in the UI
func GetGoogleAccessURL(ctx context.Context, oAuthFile string, integId string, scope ...string) (string, error) {
	config, err := GetConfig(oAuthFile, scope...)
	if err != nil {
		return "", err
	}
	return config.AuthCodeURL(integId, oauth2.AccessTypeOffline), nil
}

//GetGoogleToken retrives the token for the given scopes. This token should be stored and used for further google API calls
func GetGoogleToken(oAuthFile, code string, scope ...string) (string, error) {
	config, err := GetConfig(oAuthFile, scope...)
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

func GetConfig(oAuthFile string, scope ...string) (*oauth2.Config, error) {
	b, err := ioutil.ReadFile(oAuthFile)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read client secret file")
	}
	return google.ConfigFromJSON(b, scope...)
}

func Client(config *oauth2.Config, tokenJson string) (*http.Client, error) {
	var token oauth2.Token
	if err := json.Unmarshal([]byte(tokenJson), &token); err != nil {
		return nil, err
	}
	return config.Client(context.Background(), &token), nil
}

type DoMail interface {
	SendMail(fromName, fromEmail string, toName string, toEmail []string, subject string, body string) error
	Watch(topic string) (string, error)
}

type DoCalendar interface {
	EventCreate(calendarID, meeting Meeting) error
	Sync(calendarID string, syncToken string) (string, error)
	Watch(calendarID, channelID string) error
}

type Meeting struct {
	ID          string   `json:"id"`
	CalID       string   `json:"cal_id"`
	Summary     string   `json:"summary,omitempty"`
	HtmlLink    string   `json:"htmlLink,omitempty"`
	Description string   `json:"description,omitempty"`
	Attendees   []string `json:"attendees"`
	StartTime   string   `json:"start_time,omitempty"`
	EndTime     string   `json:"end_time,omitempty"`
	TimeZone    string   `json:"timeZone,omitempty"`
	Created     string   `json:"created,omitempty"`
	Updated     string   `json:"updated,omitempty"`
}

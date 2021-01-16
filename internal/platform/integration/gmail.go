package integration

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/mail"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

func AskGmailAccessURL(ctx context.Context, oAuthFile string) (string, error) {
	config, err := config(oAuthFile)
	if err != nil {
		return "", err
	}
	return config.AuthCodeURL("state-token", oauth2.AccessTypeOffline), nil
}

func GetToken(oAuthFile, code string) (string, error) {
	config, err := config(oAuthFile)
	if err != nil {
		return "", err
	}

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

func WatchMessage(oAuthFile, tokenJson, topicName string) (string, error) {
	config, err := config(oAuthFile)
	if err != nil {
		return "", err
	}

	var emailAddress string
	client, err := client(config, tokenJson)
	if err != nil {
		return emailAddress, err
	}

	srv, err := gmail.New(client)
	if err != nil {
		return emailAddress, err
	}

	user := "me"
	profileCall := srv.Users.GetProfile(user)
	profile, err := profileCall.Do()
	emailAddress = profile.EmailAddress
	if err != nil {
		return emailAddress, err
	}

	watchCall := srv.Users.Watch(user, &gmail.WatchRequest{
		TopicName: topicName,
	})
	_, err = watchCall.Do()
	if err != nil {
		return emailAddress, err
	}
	log.Println("started watching the user")
	return emailAddress, nil
}

func SendMail(oAuthFile, tokenJson string, user string, fromName, fromEmail string, toName, toEmail string, subject string, body string) error {
	config, err := config(oAuthFile)
	if err != nil {
		return err
	}

	client, err := client(config, tokenJson)
	if err != nil {
		return err
	}

	srv, err := gmail.New(client)
	if err != nil {
		return err
	}

	gmsg := msg(fromName, fromEmail, toName, toEmail, subject, body)
	_, err = srv.Users.Messages.Send(user, &gmsg).Do()
	if err != nil {
		return err
	}
	return nil
}

func config(oAuthFile string) (*oauth2.Config, error) {
	b, err := ioutil.ReadFile(oAuthFile)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read client secret file")
	}
	return google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
}

func client(config *oauth2.Config, tokenJson string) (*http.Client, error) {
	var token oauth2.Token
	if err := json.Unmarshal([]byte(tokenJson), &token); err != nil {
		return nil, err
	}
	return config.Client(context.Background(), &token), nil
}

func msg(fromName, fromEmail string, toName, toEmail string, subject string, body string) gmail.Message {
	from := mail.Address{fromName, fromEmail}
	to := mail.Address{toName, toEmail}

	header := make(map[string]string)
	header["From"] = from.String()
	header["To"] = to.String()
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	var msg string
	for k, v := range header {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += "\r\n" + body

	return gmail.Message{
		Raw: base64.RawURLEncoding.EncodeToString([]byte(msg)),
	}
}

type Code struct {
	Code string `json:"code"`
}

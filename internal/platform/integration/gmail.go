package integration

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

func AskGmailAccessURL(ctx context.Context, oAuthFile string) (string, error) {
	config, err := getConfig(oAuthFile)
	if err != nil {
		return "", err
	}
	return config.AuthCodeURL("state-token", oauth2.AccessTypeOffline), nil
}

func GetToken(oAuthFile, code string) (string, error) {
	config, err := getConfig(oAuthFile)
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
	config, err := getConfig(oAuthFile)
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

func getConfig(oAuthFile string) (*oauth2.Config, error) {
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

type Code struct {
	Code string `json:"code"`
}

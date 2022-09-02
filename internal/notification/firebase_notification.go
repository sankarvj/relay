package notification

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

type FirebaseNotification struct {
	AccountID string
	UserID    string
	Subject   string
	Body      string
	SDKPath   string
}

func (fbNotif FirebaseNotification) Send(ctx context.Context, notifType NotificationType, db *sqlx.DB) error {
	clients, err := RetrieveClients(ctx, fbNotif.AccountID, fbNotif.UserID, db)
	if err != nil {
		return err
	}
	for _, client := range clients {
		err = FirebaseSend(fbNotif.Subject, fbNotif.Body, client.DeviceToken, fbNotif.SDKPath)
		if err != nil {
			log.Println("*> expected error client err on firebaseSend: ", err)
			//delete token here
			DeleteClient(ctx, client.AccountID, client.UserID, client.DeviceToken, db)
			continue
		} else {
			log.Println("*> message delivered successfully")
		}
	}
	return nil
}

func FirebaseSend(subject, body string, registrationToken, adminSDKPath string) error {
	ctx := context.Background()
	opt := option.WithCredentialsFile(adminSDKPath)
	config := &firebase.Config{ProjectID: "relay-70013"}
	// Initialize default app
	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		log.Println("error getting firebase app", err)
		return err
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		log.Println("error getting Messaging client", err)
		return err
	}

	// See documentation on defining a message payload.
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: subject,
			Body:  body,
		},
		Token: registrationToken,
	}

	// Send a message to the device corresponding to the provided
	// registration token.
	response, err := client.Send(ctx, message)
	if err != nil {
		log.Println("error client.Send", err)
		return err
	}
	// Response is a message ID string.
	fmt.Println("Successfully sent message:", response)
	return nil
}

const firebaseScope = "https://www.googleapis.com/auth/firebase.messaging"

type tokenProvider struct {
	tokenSource oauth2.TokenSource
}

// newTokenProvider function to get token for fcm-send
func NewTokenProvider(credentialsLocation string) (*tokenProvider, error) {
	jsonKey, err := ioutil.ReadFile(credentialsLocation)
	if err != nil {
		return nil, errors.New("fcm: failed to read credentials file at: " + credentialsLocation)
	}
	cfg, err := google.JWTConfigFromJSON(jsonKey, firebaseScope)
	if err != nil {
		return nil, errors.New("fcm: failed to get JWT config for the firebase.messaging scope")
	}
	ts := cfg.TokenSource(context.Background())
	return &tokenProvider{
		tokenSource: ts,
	}, nil
}

func (src *tokenProvider) Token() (string, error) {
	token, err := src.tokenSource.Token()
	if err != nil {
		return "", errors.New("fcm: failed to generate Bearer token")
	}
	return token.AccessToken, nil
}

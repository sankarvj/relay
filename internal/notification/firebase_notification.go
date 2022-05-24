package notification

import (
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/jmoiron/sqlx"
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
		err = firebaseSend(client.DeviceToken, fbNotif.SDKPath)
		if err != nil {
			log.Println("error firebaseSend: ", err)
			//delete token here
			continue
		} else {
			log.Println("message delivered successfully")
		}
	}

	return nil
}

func firebaseSend(registrationToken, adminSDKPath string) error {
	ctx := context.Background()
	opt := option.WithCredentialsFile(adminSDKPath)
	config := &firebase.Config{ProjectID: "relay-70013"}
	// Initialize default app
	// config := &firebase.Config{ProjectID: "relay-94b69"}
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
		Data: map[string]string{
			"score": "850",
			"time":  "2:45",
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

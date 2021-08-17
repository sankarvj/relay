package handlers

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/aws"
)

// AwsSnsSubscription provides support for subscribtion/message .
type AwsSnsSubscription struct {
	db *sqlx.DB
}

const subConfrmType = "SubscriptionConfirmation"
const notificationType = "Notification"

//Create confirms SNS topic subscription
func (ass *AwsSnsSubscription) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	accountkey := params["accountkey"]
	productkey := params["productkey"]

	subscription, err := getSusbscription(r.Body)
	if err != nil {
		return errors.Wrap(err, "unable to decode subscription while the message is passed")
	}

	log.Printf("internal.handlers.aws : accountkey: %s : productkey: %s : subscription type : %s\n", accountkey, productkey, subscription.Type)

	if subscription.Type == subConfrmType {
		go confirmSubscription(subscription.SubscribeURL)
	} else if subscription.Type == notificationType {
		message := getMessage(subscription.Message)
		log.Println("internal.handlers.aws recieved this message : ", message)
	}
	return nil
}

func confirmSubscription(subcribeURL string) {
	response, err := http.Get(subcribeURL)
	if err != nil {
		log.Println("unexpected error occurred. Unbale to confirm subscriptions")
	} else {
		log.Printf("internal.handlers.aws subscription confirmed sucessfully. %d\n", response.StatusCode)
	}
}

func getSusbscription(reqBody io.Reader) (*aws.Subscription, error) {
	var subscription aws.Subscription
	body, err := ioutil.ReadAll(reqBody)
	log.Println("internal.handlers.aws body:", string(body))
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &subscription)
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

func getMessage(messageBody string) aws.Message {
	var message aws.Message
	err := json.Unmarshal([]byte(messageBody), &message)
	if err != nil {
		log.Println("unexpected error occurred. unable to unmarshal message body. error:", err)
		return message
	}
	return message
}

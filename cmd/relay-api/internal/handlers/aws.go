package handlers

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/aws"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration/email"
)

// AwsSnsSubscription provides support for subscribtion/message .
type AwsSnsSubscription struct {
	db    *sqlx.DB
	rPool *redis.Pool
}

const subConfrmType = "SubscriptionConfirmation"
const notificationType = "Notification"

//Create confirms SNS topic subscription
func (ass *AwsSnsSubscription) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	accountkey := params["accountkey"]
	productkey := params["productkey"]

	body, err := getBody(r.Body)
	if err != nil {
		return errors.Wrap(err, "unable to decode body when the reader is passed")
	}

	subscription, err := getSusbscription(body)
	if err != nil {
		return errors.Wrap(err, "unable to decode subscription when the body is passed")
	}

	log.Printf("internal.handlers.aws : accountkey: %s : productkey: %s : subscription type : %s\n", accountkey, productkey, subscription.Type)

	if subscription.Type == subConfrmType {
		go confirmSubscription(subscription.SubscribeURL)
	} else if subscription.Type == notificationType {
		message := getMessage(subscription.Message)
		log.Println("internal.handlers.aws recieved this message : ", message)
	} else {
		mb, err := getMailBody(body)
		if err != nil {
			return errors.Wrap(err, "unable to decode mailbody while the body is passed")
		}
		return receiveSESEmail(ctx, mb, ass.db, ass.rPool)
	}
	return nil
}

func getMailBody(body []byte) (email.MailBody, error) {
	var mailBody email.MailBody
	err := json.Unmarshal([]byte(body), &mailBody)
	if err != nil {
		return mailBody, err
	}
	return mailBody, nil
}

func confirmSubscription(subcribeURL string) {
	response, err := http.Get(subcribeURL)
	if err != nil {
		log.Println("unexpected error occurred. Unbale to confirm subscriptions")
	} else {
		log.Printf("internal.handlers.aws subscription confirmed sucessfully. %d\n", response.StatusCode)
	}
}

func getBody(reqBody io.Reader) ([]byte, error) {
	body, err := ioutil.ReadAll(reqBody)
	log.Println("internal.handlers.aws body:", string(body))
	if err != nil {
		return nil, err
	}
	return body, nil
}

func getSusbscription(body []byte) (*aws.Subscription, error) {
	var subscription aws.Subscription
	err := json.Unmarshal(body, &subscription)
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

func getMessage(messageBody string) aws.Message {
	var message aws.Message
	err := json.Unmarshal([]byte(messageBody), &message)
	if err != nil {
		return message
	}
	return message
}

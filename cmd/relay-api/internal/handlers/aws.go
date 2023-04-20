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
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration/email"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
)

var ErrForbidden = web.NewRequestError(
	errors.New("AWS SNS not authorized with valid keys"),
	http.StatusForbidden,
)

// AwsSnsSubscription provides support for subscribtion/message .
type AwsSnsSubscription struct {
	db            *sqlx.DB
	sdb           *database.SecDB
	authenticator *auth.Authenticator
}

const subConfrmType = "SubscriptionConfirmation"
const notificationType = "Notification"
const deliveryType = "Delivery"

// Create confirms SNS topic subscription
func (ass *AwsSnsSubscription) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	accountkey := params["accountkey"]
	productkey := params["productkey"]

	//TODO: Need to remove the hardcoded keys from here.
	if accountkey != "8ojTPK9k1b" && productkey != "3cxGvqPwfT" {
		return ErrForbidden
	}

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
	} else if subscription.Type == deliveryType {
		log.Println("internal.handlers.aws recieved delivery confirmation : ", subscription.Message)
	} else {
		mb, err := getMailBody(body)
		if err != nil {
			return errors.Wrap(err, "unable to decode mailbody while the body is passed")
		}
		return receiveSESEmail(ctx, mb, ass.db, ass.sdb, ass.authenticator.FireBaseAdminSDK)
	}
	return nil
}

func (ass *AwsSnsSubscription) ManageIncidents(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	accountkey := params["account_key"]

	claims, err := ass.authenticator.ParseClaims(accountkey)
	if err != nil {
		return web.NewRequestError(err, http.StatusUnauthorized)
	}

	accountID := claims.Subject

	log.Println("accountkey ", accountkey)
	log.Println("claims ", claims)
	log.Println("claims subject", claims.Subject)

	body, err := getBody(r.Body)
	if err != nil {
		return errors.Wrap(err, "unable to decode body when the reader is passed")
	}

	subscription, err := getSusbscription(body)
	if err != nil {
		return errors.Wrap(err, "unable to decode subscription when the body is passed")
	}

	log.Printf("internal.handlers.aws : accountkey: %s  : subscription type : %s\n", accountkey, subscription.Type)
	log.Printf("internal.handlers.aws raw message : %+v\n", subscription)
	log.Printf("internal.handlers.aws body body : %+v\n", string(body))

	if subscription.Type == subConfrmType {
		go confirmSubscription(subscription.SubscribeURL)
	} else if subscription.Type == notificationType {
		message := getMessage(subscription.Message)
		err = saveAlertMessage(ctx, accountID, message, ass.db, ass.sdb, ass.authenticator.FireBaseAdminSDK)
		if err != nil {
			return errors.Wrap(err, "cannot save alert message from aws")
		}
		log.Println("internal.handlers.aws recieved this message : ", message)
	} else if subscription.Type == deliveryType {
		log.Println("internal.handlers.aws recieved delivery confirmation : ", subscription.Message)
	} else {
		message := getMessage(subscription.Message)
		log.Println("internal.handlers.aws recieved this message : ", message)
		if message.AlarmName == "" {
			mb, err := getMailBody(body)
			if err != nil {
				return errors.Wrap(err, "unable to decode mailbody while the body is passed")
			}
			return receiveSESEmail(ctx, mb, ass.db, ass.sdb, ass.authenticator.FireBaseAdminSDK)
		} else {
			err = saveAlertMessage(ctx, accountID, message, ass.db, ass.sdb, ass.authenticator.FireBaseAdminSDK)
			if err != nil {
				return errors.Wrap(err, "cannot save alert message from aws")
			}
		}
	}
	return nil
}

func saveAlertMessage(ctx context.Context, accountID string, msg aws.Message, db *sqlx.DB, sdb *database.SecDB, fbSDKPath string) error {
	descJsonStr := msg.AlarmDescription
	fieldsMap := map[string]interface{}{}
	err := json.Unmarshal([]byte(descJsonStr), &fieldsMap)
	if err != nil {
		return err
	}
	return aws.SaveAlert(ctx, accountID, fieldsMap, db, sdb, fbSDKPath)
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
		log.Println("***> unexpected error occurred in internal.handlers.aws. when confirming subscriptions", err)
	} else {
		log.Printf("internal.handlers.aws subscription confirmed sucessfully. %d\n", response.StatusCode)
	}
}

func getBody(reqBody io.Reader) ([]byte, error) {
	body, err := ioutil.ReadAll(reqBody)
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

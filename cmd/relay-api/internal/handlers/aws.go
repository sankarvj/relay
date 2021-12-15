package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/mail"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/aws"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration/email"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/schema"
)

// AwsSnsSubscription provides support for subscribtion/message .
type AwsSnsSubscription struct {
	db *sqlx.DB
}

const subConfrmType = "SubscriptionConfirmation"
const notificationType = "Notification"
const deliveryType = "Delivery"

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
		data, err := base64.StdEncoding.DecodeString(mb.Content)
		if err != nil {
			log.Fatal("error:", err)
		}
		m, err := mail.ReadMessage(strings.NewReader(string(data)))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("mmmmmmmmmmmm %+v", m)

		header := m.Header
		fmt.Println("Date:", header.Get("Date"))
		fmt.Println("From:", header.Get("From"))
		fmt.Println("To:", header.Get("To"))
		fmt.Println("Subject:", header.Get("Subject"))
		fmt.Println("MessageID:", header.Get("Message-ID"))
		fmt.Println("References:", header.Get("References"))
		fmt.Println("ReplyTo:", header.Get("In-Reply-To"))

		body, err := ioutil.ReadAll(m.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s", body)

		messageID := header.Get("Message-ID")
		if header.Get("References") == "" { // save as the message
			emailEntityItem := entity.EmailEntity{
				From:    []string{header.Get("From")},
				To:      []string{header.Get("To")},
				Cc:      []string{""},
				Bcc:     []string{""},
				Subject: header.Get("Subject"),
				Body:    fmt.Sprintf("{{%s}}", body),
			}

			accountID := schema.SeedAccountID                // TODO: find the actual accountID from the domain name
			teamID := "00b0205c-477d-40ad-b781-c2e3d9c19087" // TODO: find the actual teamID from the domain name
			err = entity.SaveFixedEntityItem(ctx, accountID, teamID, schema.SeedSystemUserID, entity.FixedEntityEmails, "", messageID, integration.TypeFallback, util.ConvertInterfaceToMap(emailEntityItem), ass.db)
			if err != nil {
				return err
			}
		} else { // save as a conversation

		}

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

func getMailBody(body []byte) (email.MailBody, error) {
	var mailBody email.MailBody
	err := json.Unmarshal([]byte(body), &mailBody)
	if err != nil {
		return mailBody, err
	}
	return mailBody, nil
}

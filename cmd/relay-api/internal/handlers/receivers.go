package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/discovery"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	integ "gitlab.com/vjsideprojects/relay/internal/integration"
	"gitlab.com/vjsideprojects/relay/internal/integration/calendar"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
)

func (g *Integration) Act(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	integrationID := params["integration_id"]
	actionID := params["action_id"]
	accountID := params["account_id"]
	var actionPayload integ.ActionPayload
	if err := web.Decode(r, &actionPayload); err != nil {
		return errors.Wrap(err, "cannot parse the actionPayload")
	}

	switch integrationID {
	case integration.TypeGmail:
		return web.Respond(ctx, w, "FAILURE", http.StatusNotImplemented)
	case integration.TypeGoogleCalendar:
		c := calendar.Calendar{}
		c.Act(ctx, accountID, actionID, actionPayload, g.db)
	default:
		return web.Respond(ctx, w, "FAILURE", http.StatusNotImplemented)
	}
	return web.Respond(ctx, w, "SUCCESS", http.StatusOK)
}

func (g *Integration) Notifications(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	log.Printf("handlers.receivers: notifications received with body %s\n", r.Body)
	return web.Respond(ctx, w, "SUCCESS", http.StatusOK)
}

func (g *Integration) ReceiveEmail(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	var pushMsgPayload PushMsgPayload
	if err := web.Decode(r, &pushMsgPayload); err != nil {
		return errors.Wrap(err, "parsing push message payload")
	}
	raw, err := base64.StdEncoding.DecodeString(pushMsgPayload.Message.Data)
	if err != nil {
		return errors.Wrap(err, "decoding data in push message payload")
	}
	var data Data
	err = json.Unmarshal(raw, &data)
	if err != nil {
		return errors.Wrap(err, "unmarshal data from the data bytes")
	}

	sub, err := discovery.Retrieve(ctx, data.EmailAddress, g.db)
	if err != nil {
		if err == discovery.ErrDiscoveryEmpty { //means we don't want to listen to that mailbox
			//TODO call stop here.
			log.Println("internal.handlers.receivers silently killing the unwanted messages.")
			return web.Respond(ctx, w, "SUCCESS", http.StatusOK)
		}
		return err

	}

	valueAddedConfigFields, _, err := entity.RetrieveFixedItem(ctx, sub.AccountID, sub.EntityID, sub.ItemID, g.db)
	if err != nil {
		return err
	}
	var emailConfigEntityItem entity.EmailConfigEntity
	err = entity.ParseFixedEntity(valueAddedConfigFields, &emailConfigEntityItem)
	if err != nil {
		return err
	}
	//integration.History(g.authenticator.GoogleClientSecret, emailConfigEntityItem.APIKey, data.EmailAddress, 1709032)

	return web.Respond(ctx, w, "SUCCESS", http.StatusOK)
}

type PushMsgPayload struct {
	Message struct {
		Data         string    `json:"data"`
		MessageID    string    `json:"message_id"`
		MessageId    string    `json:"messageId"`
		PublishTime  time.Time `json:"publish_time"`
		PublishTime1 time.Time `json:"publishTime"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

type Data struct {
	EmailAddress string `json:"emailAddress"`
	HistoryID    uint64 `json:"historyId"`
}

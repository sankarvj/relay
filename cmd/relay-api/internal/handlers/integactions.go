package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/calendar"
	"gitlab.com/vjsideprojects/relay/internal/discovery"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/user"
)

func (g *Integration) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	integrationID := params["integration_id"]
	actionID := params["action_id"]
	accountID := params["account_id"]
	var actionPayload integration.ActionPayload
	if err := web.Decode(r, &actionPayload); err != nil {
		return errors.Wrap(err, "cannot parse the actionPayload")
	}

	log.Println("actionID --> ", actionID)

	switch integrationID {
	case integration.TypeGmail:

	case integration.TypeGoogleCalendar:
		calendar.CreateEvent(ctx, accountID, actionPayload, g.db)
	default:
		return web.Respond(ctx, w, "FAILURE", http.StatusNotImplemented)
	}
	return web.Respond(ctx, w, "SUCCESS", http.StatusOK)
}

func (g *Integration) DailyWatch(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return err
	}

	emailsConfigEntity, err := entity.RetrieveFixedEntity(ctx, g.db, params["account_id"], entity.FixedEntityEmailConfig)
	if err != nil {
		return err
	}

	fields, err := emailsConfigEntity.Fields()
	if err != nil {
		return err
	}

	emailConfigs, err := item.UserEntityItems(ctx, emailsConfigEntity.ID, currentUserID, g.db)
	if err != nil {
		return err
	}

	for _, emailConfig := range emailConfigs {
		valueAddedConfigFields := entity.ValueAddFields(fields, emailConfig.Fields())
		var emailConfigEntityItem entity.EmailConfigEntity
		err = entity.ParseFixedEntity(valueAddedConfigFields, &emailConfigEntityItem)
		if err != nil {
			return err
		}
		if strings.HasSuffix(emailConfigEntityItem.Domain, integration.DomainGMail) {
			_, err = integration.WatchMessage(g.authenticator.GoogleClientSecret, emailConfigEntityItem.APIKey, g.publisher.Topic)
			if err != nil {
				return err
			}
		}

	}
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
		if err == discovery.ErrDiscoveryFailed { //means we don't want to listen to that mailbox
			//TODO call stop here.
			log.Println("Silently Killing The Unwanted Messages...")
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
	integration.History(g.authenticator.GoogleClientSecret, emailConfigEntityItem.APIKey, data.EmailAddress, 1709032)

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

package handlers

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/discovery"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration"
	"gitlab.com/vjsideprojects/relay/internal/platform/pubsub"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
	"golang.org/x/net/context"
)

var (
	// TokenNotFound is used when a specific token does not exist.
	TokenNotFound = errors.New("Token not found")
)

// Integration represents the data needed for the third party integration details.
type Integration struct {
	db            *sqlx.DB
	authenticator *auth.Authenticator
	publisher     *pubsub.Publisher
}

func (g *Integration) AccessIntegration(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "internal.email.access")
	defer span.End()
	integrationID := params["integration_id"]
	var (
		authURL string
		err     error
	)
	switch integrationID {
	case integration.TypeGmail:
		authURL, err = integration.AskGmailAccessURL(ctx, g.authenticator.GoogleClientSecret)
	}
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, authURL, http.StatusOK)
}

func (g *Integration) SaveIntegration(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.email.saveIntegration")
	defer span.End()

	var (
		emailAddress string
		tokenJson    string
		err          error
	)

	integrationID := params["integration_id"]
	accountID := params["account_id"]
	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return err
	}

	var code Code
	if err := web.Decode(r, &code); err != nil {
		return errors.Wrap(err, "cannot parse the code")
	}

	switch integrationID {
	case integration.TypeGmail:
		tokenJson, err = integration.GetToken(g.authenticator.GoogleClientSecret, code.Code)
		if err != nil {
			return err
		}
		emailAddress, err = integration.WatchMessage(g.authenticator.GoogleClientSecret, tokenJson, g.publisher.Topic)
		if err != nil {
			return err
		}
	}

	_, err = entity.SaveEmailIntegration(ctx, accountID, currentUserID, integration.DomainGMail, tokenJson, emailAddress, g.db)
	if err != nil {
		return err
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
	integration.History(g.authenticator.GoogleClientSecret, emailConfigEntityItem.APIKey, data.EmailAddress, data.HistoryID)

	return web.Respond(ctx, w, "SUCCESS", http.StatusOK)
}

type Code struct {
	Code string `json:"code"`
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

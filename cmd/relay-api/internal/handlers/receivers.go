package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DusanKasan/parsemail"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	conv "gitlab.com/vjsideprojects/relay/internal/conversation"
	"gitlab.com/vjsideprojects/relay/internal/discovery"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	integ "gitlab.com/vjsideprojects/relay/internal/integration"
	"gitlab.com/vjsideprojects/relay/internal/integration/calendar"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration/email"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/schema"
)

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

/***

					GOOGLE RECEIVERS - GMAIL/CALENDAR API - GCONSOLE TOPIC

***/
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

/***

					GOOGLE RECEIVERS - GMAIL API - GCONSOLE TOPIC

***/
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

	log.Printf("data.raw raw raw---> %+v", pushMsgPayload)
	log.Println("data.EmailAddress---> ", data.EmailAddress)
	log.Println("data.HistoryID---> ", data.HistoryID)

	sub, err := discovery.Retrieve(ctx, "", "", data.EmailAddress, g.db)
	if err != nil {
		if err == discovery.ErrDiscoveryEmpty { //means we don't want to listen to that mailbox
			//TODO call stop here.
			log.Println("internal.handlers.receivers silently killing the unwanted messages.")
			return web.Respond(ctx, w, "SUCCESS", http.StatusOK)
		}
		return err

	}

	valueAddedConfigFields, updaterFunc, err := entity.RetrieveFixedItem(ctx, sub.AccountID, sub.EntityID, sub.ItemID, g.db)
	if err != nil {
		return err
	}
	var emailConfigEntityItem entity.EmailConfigEntity
	err = entity.ParseFixedEntity(valueAddedConfigFields, &emailConfigEntityItem)
	if err != nil {
		return err
	}
	hisID, _ := strconv.ParseUint(emailConfigEntityItem.HistoryID, 10, 64)
	if hisID != 0 {
		log.Println("Calling History For ", hisID)
		_, err = email.History(g.authenticator.GoogleClientSecret, emailConfigEntityItem.APIKey, data.EmailAddress, data.HistoryID)
		if err != nil {
			log.Println("Err -> ", err)
			return err
		}
	}
	//save the new history ID
	emailConfigEntityItem.HistoryID = strconv.FormatUint(data.HistoryID, 10)
	updaterFunc(ctx, emailConfigEntityItem, g.db)

	return web.Respond(ctx, w, "SUCCESS", http.StatusOK)
}

/***

					AWS RECEIVERS

***/

func receiveSESEmail(ctx context.Context, mb email.MailBody, db *sqlx.DB, rp *redis.Pool) error {
	data, err := base64.StdEncoding.DecodeString(mb.Content)
	if err != nil {
		return err
	}

	email, err := parsemail.Parse(strings.NewReader(string(data))) // returns Email struct and error
	if err != nil {
		return err
	}

	messageID := email.MessageID
	replyTo := email.InReplyTo
	references := email.References
	to := email.To[0].Address
	from := email.From[0].Address
	subject := email.Subject

	discoveryID := util.SubDomainInEmail(to)
	fmt.Printf("DiscoveryID: %s, MessageID: %s, ReplyTo: %s, References: %s", discoveryID, messageID, replyTo, references)
	var emailConfigEntityItem entity.EmailConfigEntity
	eConfigItemID, err := entity.DiscoverAnyEntityItem(ctx, "", "", discoveryID, &emailConfigEntityItem, db)
	if err != nil {
		return err
	}

	emailEntityItem := entity.EmailEntity{
		From:      []string{eConfigItemID},
		RFrom:     []string{from},
		To:        []string{to},
		Cc:        []string{},
		Bcc:       []string{},
		Contacts:  []string{},
		Companies: []string{},
		Subject:   subject,
		Body:      email.HTMLBody,
	}

	fmt.Printf("\nIncomingMessage: %+v", emailEntityItem)

	if len(references) == 0 { // save as the message
		err = saveEmailPlusConnect(ctx, emailConfigEntityItem.AccountID, emailConfigEntityItem.TeamID, messageID, emailEntityItem, db, rp)
		if err != nil {
			return errors.Wrap(err, "unable to save the email received via SES for the discoveryID")
		}
	} else { // save as a conversation for the message saved
		reference := references[0]
		fixedEmailEntity, err := entity.RetrieveFixedEntity(ctx, db, emailConfigEntityItem.AccountID, emailConfigEntityItem.TeamID, entity.FixedEntityEmails)
		if err != nil {
			return err
		}
		var emailEntityItem entity.EmailEntity
		parentItemId, err := entity.DiscoverAnyEntityItem(ctx, fixedEmailEntity.AccountID, fixedEmailEntity.ID, reference, &emailEntityItem, db)
		if err != nil {
			return err
		}
		return saveConversation(ctx, fixedEmailEntity.AccountID, fixedEmailEntity.ID, parentItemId, emailEntityItem, db)

	}
	return nil
}

func saveEmailPlusConnect(ctx context.Context, accountID, teamID, messageID string, emailEntityItem entity.EmailEntity, db *sqlx.DB, rp *redis.Pool) error {
	contacts, err := createContactIfNotExist(ctx, accountID, teamID, emailEntityItem.RFrom[0], db, rp)
	if err != nil {
		return err
	}
	//associating the contact. TODO Shall we move this to the workflow?? It is the vision
	emailEntityItem.Contacts = contacts
	//using the account and team of the emailConfig we are saving the emails inside the same acc/team
	it, err := entity.SaveFixedEntityItem(ctx, accountID, teamID, schema.SeedSystemUserID, entity.FixedEntityEmails, "received", messageID, integration.TypeMails, util.ConvertInterfaceToMap(emailEntityItem), db)
	if err != nil {
		return err
	}
	//TODO push this to stream/queue
	(&job.Job{}).EventItemCreated(it.AccountID, it.EntityID, it.ID, map[string]string{}, db, rp)
	return nil
}

func saveConversation(ctx context.Context, accountID, entityID, parentItemId string, emailEntityItem entity.EmailEntity, db *sqlx.DB) error {
	newConversation := conv.NewConversation{
		ID:        uuid.New().String(),
		AccountID: accountID,
		EntityID:  entityID,
		ItemID:    &parentItemId,
		UserID:    schema.SeedSystemUserID,
		Payload:   util.ConvertInterfaceToMap(emailEntityItem),
	}
	_, err := conv.Create(ctx, db, newConversation, time.Now())
	if err != nil {
		return errors.Wrap(err, "error inserting the conversation message to the DB")
	}
	return nil
}

func createContactIfNotExist(ctx context.Context, accountID, teamID, value string, db *sqlx.DB, rp *redis.Pool) ([]string, error) {
	e, err := entity.RetrieveFixedEntity(ctx, db, accountID, teamID, schema.SeedContactsEntityName)
	if err != nil {
		return []string{}, err
	}

	exp := fmt.Sprintf("{{%s.%s}} eq {%s}", e.ID, e.Key("email"), value)
	result, err := segment(ctx, accountID, e.ID, exp, db, rp)
	if err != nil {
		return []string{}, err
	}
	itemIds := itemIDs(result)

	if len(itemIds) == 0 {
		fields := make(map[string]interface{}, 0)
		e.FilteredFields()
		name := "System Generated"
		fields[e.Key("email")] = value
		//create a new contact here
		ni := item.NewItem{
			ID:        uuid.New().String(),
			Name:      &name,
			AccountID: accountID,
			EntityID:  e.ID,
			Fields:    fields,
		}

		it, err := item.Create(ctx, db, ni, time.Now())
		if err != nil {
			return []string{}, err
		}
		itemIds = append(itemIds, it.ID)

	}

	return util.ConvertSliceTypeRev(itemIds), nil
}

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
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/conversation"
	"gitlab.com/vjsideprojects/relay/internal/discovery"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	integ "gitlab.com/vjsideprojects/relay/internal/integration"
	"gitlab.com/vjsideprojects/relay/internal/integration/calendar"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/database/dbservice"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration/email"
	"gitlab.com/vjsideprojects/relay/internal/platform/stream"
	"gitlab.com/vjsideprojects/relay/internal/platform/tracker"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/schema"
	"gitlab.com/vjsideprojects/relay/internal/user"
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

/*
**

	GOOGLE RECEIVERS - GMAIL/CALENDAR API - GCONSOLE TOPIC

**
*/
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

/*
**

	GOOGLE RECEIVERS - GMAIL API - GCONSOLE TOPIC

**
*/
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

	sub, err := discovery.Retrieve(ctx, "", "", data.EmailAddress, g.db)
	if err != nil {
		if err == discovery.ErrDiscoveryEmpty { //means we don't want to listen to that mailbox
			//TODO call stop here.
			log.Println("*********> debug internal.handlers.receivers silently killing the unwanted messages.")
			return web.Respond(ctx, w, "SUCCESS", http.StatusOK)
		}
		return err

	}

	valueAddedConfigFields, updaterFunc, err := entity.RetrieveFixedItem(ctx, sub.AccountID, sub.EntityID, sub.ItemID, g.db, g.sdb)
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
		_, err = email.History(g.authenticator.GoogleClientSecret, emailConfigEntityItem.APIKey, data.EmailAddress, data.HistoryID)
		if err != nil {
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

func receiveSESEmail(ctx context.Context, mb email.MailBody, db *sqlx.DB, sdb *database.SecDB, fbSDKPath string) error {
	data, err := base64.StdEncoding.DecodeString(mb.Content)
	if err != nil {
		return err
	}

	email, err := parsemail.Parse(strings.NewReader(string(data))) // returns Email struct and error
	if err != nil {
		return err
	}

	replyTo := email.InReplyTo
	references := email.References
	to := email.To[0].Address
	from := email.From[0].Address
	subject := email.Subject
	messageID := util.MessageID(email.MessageID)
	uniqueDomainID := util.SubDomain(to)

	fmt.Printf("UniqueDomainID: %s \n MessageID: %s \n From: %s \n InReplyTo: %s \n References: %s \n ReplyTo: %s \n", uniqueDomainID, messageID, from, replyTo, references, email.ReplyTo)
	var emailConfigEntityItem entity.EmailConfigEntity
	eConfigItemID, err := entity.DiscoverAnyEntityItem(ctx, "", "", uniqueDomainID, &emailConfigEntityItem, db, sdb)
	if err != nil {
		return err
	}

	emailEntityItem := entity.EmailEntity{
		MessageID: messageID,
		From:      []string{eConfigItemID},
		RFrom:     []string{from},
		To:        []string{to},
		Cc:        []string{},
		Bcc:       []string{},
		Subject:   subject,
		Body:      email.HTMLBody,
	}

	tracker.EmailChan().Log("Email Received", fmt.Sprintf("Received email from domain %s", emailConfigEntityItem.Domain))

	if len(references) == 0 { // save as the first message
		source, err := creatSourceIfNotExist(ctx, emailConfigEntityItem.AccountID, emailConfigEntityItem.TeamID, emailEntityItem, db, sdb, fbSDKPath)
		if err != nil {
			return errors.Wrap(err, "problem retriving/creating the contact/employee to associate a email")
		}
		//using emailConfig's account_id/team_id to save the email entity.
		savedEmailItem, err := entity.SaveFixedEntityItem(ctx, emailConfigEntityItem.AccountID, emailConfigEntityItem.TeamID, schema.SeedSystemUserID, entity.FixedEntityEmails, "", "", "", util.ConvertInterfaceToMap(emailEntityItem), db)
		if err != nil {
			return err
		}
		//associating with the source.
		//TODO: shall we move association to the workflow?
		go job.NewJob(db, sdb, fbSDKPath).Stream(stream.NewCreteItemMessage(ctx, db, savedEmailItem.AccountID, schema.SeedSystemUserID, savedEmailItem.EntityID, savedEmailItem.ID, source))
	} else { // save as a conversation for the message saved
		reference := util.MessageID(references[0])
		parentConv, err := conversation.Retrieve(ctx, emailConfigEntityItem.AccountID, reference, db)
		if err != nil || parentConv.ItemID == nil {
			return errors.Wrap(err, "conversation thread not exist. stopping here.")
		}

		err = conversation.SaveConversation(ctx, parentConv.AccountID, parentConv.EntityID, *parentConv.ItemID, emailEntityItem, emailEntityItem.Body, conversation.TypeConvReceived, conversation.StateDelivered, db)
		if err != nil {
			return errors.Wrap(err, "unable to save conversation")
		}
	}
	return nil
}

// creatSourceIfNotExist should take contact/employee as the source.
// TODO: what happens if the both contacts/emplooyees not exist in that account?
func creatSourceIfNotExist(ctx context.Context, accountID, teamID string, emailEntityItem entity.EmailEntity, db *sqlx.DB, sdb *database.SecDB, fbSDKPath string) (map[string][]string, error) {
	e, err := entity.RetrieveFixedEntity(ctx, db, accountID, teamID, entity.FixedEntityContacts)
	if err != nil {
		return nil, err
	}

	contacts, err := createContactIfNotExist(ctx, accountID, e, emailEntityItem.RFrom[0], db, sdb, fbSDKPath)
	if err != nil {
		tracker.ErrorChan().Log("Bug Found", fmt.Sprintf("Could not find contacts entity on account `%s`", accountID))
		return nil, err
	}
	source := map[string][]string{e.ID: contacts}

	return source, nil
}

func createContactIfNotExist(ctx context.Context, accountID string, e entity.Entity, value string, db *sqlx.DB, sdb *database.SecDB, fbSDKPath string) ([]string, error) {
	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil && err != user.ErrNotFound {
		return []string{}, err
	} else if err == user.ErrNotFound {
		currentUserID = schema.SeedSystemUserID
	}

	exp := fmt.Sprintf("{{%s.%s}} eq {%s}", e.ID, e.Key("email"), value)
	//TODO: segment call doesn't need the count. But it is executing count query in the call. Shall we stop it?

	conditionFields, err := makeConditionsFromExp(ctx, accountID, e.ID, exp, db, sdb)
	if err != nil {
		return nil, err
	}
	itemIds := dbservice.NewDBservice(dbservice.Spider, db, sdb).Search3(ctx, accountID, e.ID, conditionFields)

	if len(itemIds) == 0 {
		fields := make(map[string]interface{}, 0)
		name := "System Generated"
		fields[e.Key("first_name")] = util.NameInEmail(value)
		fields[e.Key("email")] = value
		//create a new contact here
		ni := item.NewItem{
			ID:        uuid.New().String(),
			Name:      &name,
			AccountID: accountID,
			UserID:    &currentUserID,
			EntityID:  e.ID,
			Fields:    fields,
		}

		it, err := createAndPublish(ctx, currentUserID, ni, db, sdb, fbSDKPath)
		if err != nil {
			return []string{}, err
		}
		itemIds = append(itemIds, it.ID)

	}

	return util.ConvertSliceTypeRev(itemIds), nil
}

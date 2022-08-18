package email

import (
	"context"
	"strings"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration/email"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/user"
)

func Act(actionID string) {

}

func DailyWatch(ctx context.Context, accountID, teamID, oAuthFile, topic string, db *sqlx.DB) error {
	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return err
	}

	ec, err := entity.RetrieveFixedEntity(ctx, db, accountID, teamID, entity.FixedEntityEmailConfig)
	if err != nil {
		return err
	}

	emailConfigs, err := item.UserEntityItems(ctx, ec.ID, currentUserID, db)
	if err != nil {
		return err
	}

	for _, emailConfig := range emailConfigs {
		var emailConfigEntityItem entity.EmailConfigEntity
		err = entity.ParseFixedEntity(ec.ValueAdd(emailConfig.Fields()), &emailConfigEntityItem)
		if err != nil {
			return err
		}
		if strings.HasSuffix(emailConfigEntityItem.Domain, integration.DomainGMail) {
			g := email.Gmail{OAuthFile: emailConfigEntityItem.Domain, TokenJson: emailConfigEntityItem.APIKey}
			_, err = g.Watch(topic)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func SendMail(ctx context.Context, accountID, entityID, itemID string, valueAddedMailFields []entity.Field, replyToID string, db *sqlx.DB) (*string, error) {

	namedFieldsObj := entity.NamedFieldsObjMap(valueAddedMailFields)

	fromField := namedFieldsObj["from"]
	toField := namedFieldsObj["to"]
	fromFieldValue := fromField.Value.([]interface{})[0].(string)
	subject := namedFieldsObj["subject"].Value.(string)
	body := namedFieldsObj["body"].Value.(string)

	//fetch e-mail integration config id from the from field of the mail
	var emailConfigEntityItem entity.EmailConfigEntity
	_, err := entity.RetrieveUnmarshalledItem(ctx, accountID, fromField.RefID, fromFieldValue, &emailConfigEntityItem, db)
	if err != nil {
		return nil, err
	}

	var e email.Email
	if emailConfigEntityItem.Domain == "mailgun.org" {
		e = email.MailGun{Domain: emailConfigEntityItem.Domain, TokenJson: emailConfigEntityItem.APIKey, ReplyTo: replyToID}
	} else if emailConfigEntityItem.Domain == "google.com" {
		e = email.Gmail{OAuthFile: "config/dev/google-apps-client-secret.json", TokenJson: emailConfigEntityItem.APIKey, ReplyTo: replyToID}
	} else if emailConfigEntityItem.Domain == "base_inbox.com" {
		e = email.SESMail{Domain: "", ReplyTo: replyToID}
	}

	messageID, err := e.SendMail("", emailConfigEntityItem.Email, "", util.ConvertSliceTypeRev(toField.Value.([]interface{})), subject, body)
	if err != nil {
		return nil, err
	}

	return messageID, nil
}

func Destruct(ctx context.Context, accountID, entityID, itemID string, db *sqlx.DB) error {

	var emailConfigEntityItem entity.EmailConfigEntity
	_, err := entity.RetrieveUnmarshalledItem(ctx, accountID, entityID, itemID, &emailConfigEntityItem, db)
	if err != nil {
		return err
	}

	var e email.Email
	if emailConfigEntityItem.Domain == "mailgun.org" {
		e = email.MailGun{Domain: emailConfigEntityItem.Domain, TokenJson: emailConfigEntityItem.APIKey}
	} else if emailConfigEntityItem.Domain == "google.com" {
		e = email.Gmail{OAuthFile: "config/dev/google-apps-client-secret.json", TokenJson: emailConfigEntityItem.APIKey}
	} else if emailConfigEntityItem.Domain == "base_inbox.com" {
		e = email.SESMail{Domain: ""}
	}
	return e.Stop("me")
}

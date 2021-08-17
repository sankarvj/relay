package email

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/discovery"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration/email"
	"gitlab.com/vjsideprojects/relay/internal/user"
)

func Act(actionID string) {

}

func DailyWatch(ctx context.Context, accountID, oAuthFile, topic string, db *sqlx.DB) error {
	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return err
	}

	emailsConfigEntity, err := entity.RetrieveFixedEntity(ctx, db, accountID, entity.FixedEntityEmailConfig)
	if err != nil {
		return err
	}

	emailConfigs, err := item.UserEntityItems(ctx, emailsConfigEntity.ID, currentUserID, db)
	if err != nil {
		return err
	}

	for _, emailConfig := range emailConfigs {
		var emailConfigEntityItem entity.EmailConfigEntity
		err = entity.ParseFixedEntity(emailsConfigEntity.ValueAdd(emailConfig.Fields()), &emailConfigEntityItem)
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

func SendMail(ctx context.Context, accountID, entityID, itemID string, valueAddedMailFields []entity.Field, db *sqlx.DB) error {

	namedFieldsObj := entity.NamedFieldsObjMap(valueAddedMailFields)
	fromField := namedFieldsObj["from"]
	toField := namedFieldsObj["to"]

	to := toField.ChoicesValues()
	fromFieldValue := fromField.Value.([]interface{})[0].(string)
	subject := namedFieldsObj["subject"].Value.(string)
	body := namedFieldsObj["body"].Value.(string)

	//fetch e-mail integration config id from the from field of the mail
	valueAddedConfigFields, _, err := entity.RetrieveFixedItem(ctx, accountID, fromField.RefID, fromFieldValue, db)
	if err != nil {
		return err
	}

	var emailConfigEntityItem entity.EmailConfigEntity
	err = entity.ParseFixedEntity(valueAddedConfigFields, &emailConfigEntityItem)
	if err != nil {
		return err
	}

	log.Printf("integration.email send email params : from: %+v subject: %+v body: %+v to: %+v ", emailConfigEntityItem.Email, subject, body, to)

	g := email.Gmail{OAuthFile: emailConfigEntityItem.Domain, TokenJson: emailConfigEntityItem.APIKey}

	threadID, err := g.SendMail("", emailConfigEntityItem.Email, "", to, subject, body)
	if err != nil {
		return err
	}

	ns := discovery.NewDiscovery{
		ID:        *threadID,
		AccountID: accountID,
		EntityID:  entityID,
		ItemID:    itemID,
	}

	_, err = discovery.Create(ctx, db, ns, time.Now())
	if err != nil {
		return err
	}

	return nil
}

package email

import (
	"context"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/discovery"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration"
)

func SendMail(ctx context.Context, accountID, entityID, itemID string, valueAddedMailFields []entity.Field, db *sqlx.DB) error {

	namedFieldsObj := entity.NamedFieldsObjMap(valueAddedMailFields)
	fromField := namedFieldsObj["from"]
	fromFieldValue := fromField.Value.([]interface{})[0].(string)

	toField := namedFieldsObj["to"]
	to := toField.DisplayValues()
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

	log.Printf("from --> %+v", emailConfigEntityItem.Email)
	log.Printf("subject --> %+v", subject)
	log.Printf("body --> %+v", body)
	log.Printf("toField --> %+v", to)

	threadID, err := integration.SendEmail(emailConfigEntityItem.Domain, emailConfigEntityItem.APIKey, emailConfigEntityItem.Email, to, subject, body)
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

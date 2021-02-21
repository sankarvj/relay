package job

import (
	"context"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration"
	"gitlab.com/vjsideprojects/relay/internal/reference"
)

func sendMail(ctx context.Context, accountID, entityID, itemID string, fields []entity.Field, mailFields map[string]interface{}, db *sqlx.DB) error {
	valueAddedMailFields := entity.ValueAddFields(fields, mailFields)
	reference.UpdateChoicesWrapper(ctx, db, valueAddedMailFields)
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

	return integration.SendEmail(emailConfigEntityItem.Domain, emailConfigEntityItem.APIKey, emailConfigEntityItem.Email, to, subject, body)
}

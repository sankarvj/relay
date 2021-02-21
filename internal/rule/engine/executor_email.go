package engine

import (
	"context"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func executeEmail(ctx context.Context, db *sqlx.DB, n node.Node) error {
	mailFields, err := mergeActualsWithActor(ctx, db, n.AccountID, n.ActorID, n.ActualsItemID())
	if err != nil {
		return err
	}
	//fetch e-mail integration config from the from field of the mail
	namedFieldsObj := entity.NamedFieldsObjMap(mailFields)
	fromField := namedFieldsObj["from"]
	fromFieldValue := fromField.Value.([]interface{})[0].(string)

	mailConfigFields, err := mergeActualsWithActor(ctx, db, n.AccountID, fromField.RefID, fromFieldValue)
	if err != nil {
		return err
	}

	var emailConfigEntityItem entity.EmailConfigEntity
	err = entity.ParseFixedEntity(mailConfigFields, &emailConfigEntityItem)
	if err != nil {
		return err
	}

	var emailEntityItem entity.EmailEntity
	err = entity.ParseFixedEntity(mailFields, &emailEntityItem)
	if err != nil {
		return err
	}

	//get config
	variables := n.VariablesMap()
	emailEntityItem.Body = RunExpRenderer(ctx, db, n.AccountID, emailEntityItem.Body, variables)
	tos := []string{}
	for _, to := range emailEntityItem.To {
		tos = append(tos, RunExpRenderer(ctx, db, n.AccountID, to, variables))
	}
	emailEntityItem.To = tos
	emailEntityItem.Subject = RunExpRenderer(ctx, db, n.AccountID, emailEntityItem.Subject, variables)

	return integration.SendEmail(emailConfigEntityItem.Domain, emailConfigEntityItem.APIKey, emailConfigEntityItem.Email, emailEntityItem.To, emailEntityItem.Subject, emailEntityItem.Body)
}

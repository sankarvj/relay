package engine

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/integration/email"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func (eng *Engine) executeEmailMayBeRemoved(ctx context.Context, db *sqlx.DB, n node.Node) error {
	mailFields, err := valueAdd(ctx, db, n.AccountID, n.ActorID, n.ActualsItemID())
	if err != nil {
		return err
	}
	namedFieldsObj := entity.NamedFieldsObjMap(mailFields)
	var (
		subject string
		body    string
		to      []interface{}
	)
	if namedFieldsObj["subject"].Value != nil {
		subject = namedFieldsObj["subject"].Value.(string)
	}
	if namedFieldsObj["body"].Value != nil {
		body = namedFieldsObj["body"].Value.(string)
	}
	if namedFieldsObj["to"].Value != nil {
		to = namedFieldsObj["to"].Value.([]interface{})
	}

	variables := n.VariablesMap()

	subject = eng.RunExpRenderer(ctx, db, n.AccountID, subject, variables)
	body = eng.RunExpRenderer(ctx, db, n.AccountID, body, variables)

	log.Printf("mailFields --> %+v", mailFields)

	for i := 0; i < len(mailFields); i++ {
		switch mailFields[i].Name {
		case "subject":
			mailFields[i].Value = subject
		case "body":
			mailFields[i].Value = body
		case "to":
			mailFields[i].Value = to
		}
	}

	return email.SendMail(ctx, n.AccountID, n.ActorID, n.ActualsItemID(), mailFields, db)
}

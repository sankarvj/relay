package engine

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/email"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func executeEmail(ctx context.Context, db *sqlx.DB, n node.Node) error {
	mailFields, err := mergeActualsWithActor(ctx, db, n.AccountID, n.ActorID, n.ActualsItemID())
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

	choices := make([]entity.Choice, 0)
	variables := n.VariablesMap()

	subject = RunExpRenderer(ctx, db, n.AccountID, subject, variables)
	body = RunExpRenderer(ctx, db, n.AccountID, body, variables)
	//Very confusing step.
	//To satisfy the SendMail func in the job we are populating the to mails in the choices.
	for _, t := range to {
		choices = append(choices, entity.Choice{DisplayValue: RunExpRenderer(ctx, db, n.AccountID, t.(string), variables)})
	}

	log.Printf("mailFields --> %+v", mailFields)

	for i := 0; i < len(mailFields); i++ {
		switch mailFields[i].Name {
		case "subject":
			mailFields[i].Value = subject
		case "body":
			mailFields[i].Value = body
		case "to":
			mailFields[i].Choices = choices
		}
	}

	return email.SendMail(ctx, n.AccountID, n.ActorID, n.ActualsItemID(), mailFields, db)
}

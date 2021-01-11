package engine

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/mailgun/mailgun-go"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func executeEmail(ctx context.Context, db *sqlx.DB, n node.Node) error {
	mailFields, err := mergeActualsWithActor(ctx, db, n.ActorID, n.ActualsMap())
	if err != nil {
		return err
	}

	namedFieldsObj := namedFieldsObjMap(mailFields)
	fromField := namedFieldsObj["from"]

	mailConfigFields, err := mergeActualsWithActor(ctx, db, fromField.RefID, map[string]string{fromField.RefID: fromField.Value.(string)})
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
	emailEntityItem.Body = RunExpRenderer(ctx, db, emailEntityItem.Body, variables)
	emailEntityItem.To = RunExpRenderer(ctx, db, emailEntityItem.To, variables)
	emailEntityItem.Subject = RunExpRenderer(ctx, db, emailEntityItem.Subject, variables)

	switch {
	case strings.HasSuffix(emailConfigEntityItem.Domain, integration.DomainMailGun):
		sendSimpleMessage(emailConfigEntityItem.Domain, emailConfigEntityItem.APIKey, emailConfigEntityItem.Email, emailEntityItem.To, emailEntityItem.Subject, emailEntityItem.Body)
	case strings.HasSuffix(emailConfigEntityItem.Domain, integration.DomainGMail):
		return errors.New("G-Mail send message not implemented yet")
	default:
		return errors.New("No e-mail client configured to send the mail template")
	}

	return nil
}

func sendSimpleMessage(domain, key, from, to, subject, body string) (string, error) {
	mg := mailgun.NewMailgun(domain, key)
	m := mg.NewMessage(
		from,
		subject,
		body,
		to,
	)
	resMsg, id, err := mg.Send(m)
	log.Println("resMsg ", resMsg)
	log.Println("resMsg id ", id)
	log.Println("resMsg err ", err)
	return id, err
}

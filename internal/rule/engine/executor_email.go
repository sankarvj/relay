package engine

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/mailgun/mailgun-go"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func executeEmail(ctx context.Context, db *sqlx.DB, entityFields []entity.Field, n node.Node) error {
	entityFields, err := fillItemFieldValues(ctx, db, entityFields, n.ActualsMap()[n.ActorID])
	if err != nil {
		return err
	}

	emailEntity, err := entity.ParseEmailEntity(namedFieldsMap(entityFields))
	if err != nil {
		log.Println("err ", err)
	}

	variables := n.VariablesMap()
	emailEntity.Body = RunRenderer(ctx, db, emailEntity.Body, variables)
	emailEntity.To = RunRenderer(ctx, db, emailEntity.To, variables)
	emailEntity.Subject = RunRenderer(ctx, db, emailEntity.Subject, variables)
	emailEntity.Sender = RunRenderer(ctx, db, emailEntity.Sender, variables)

	_, err = sendSimpleMessage(emailEntity)
	return err
}

func sendSimpleMessage(emailEntity entity.EmailEntity) (string, error) {
	mg := mailgun.NewMailgun(emailEntity.Domain, emailEntity.APIKey)
	m := mg.NewMessage(
		emailEntity.Sender,
		emailEntity.Subject,
		emailEntity.Body,
		emailEntity.To,
	)
	log.Printf("emailEntity %+v", emailEntity)
	resMsg, id, err := mg.Send(m)
	log.Println("resMsg ", resMsg)
	log.Println("resMsg id ", id)
	log.Println("resMsg err ", err)
	return id, err
}

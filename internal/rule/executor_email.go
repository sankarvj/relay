package rule

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/mailgun/mailgun-go"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func executeEmail(ctx context.Context, db *sqlx.DB, input map[string]string, entityFields []entity.Field) error {
	params := namedFieldsMap(entityFields)
	emailEntity, err := entity.ParseEmailEntity(params)
	if err != nil {
		log.Println("err ", err)
	}
	emailEntity.Body = RunRuleEngine(ctx, db, emailEntity.Body, input)
	emailEntity.To = RunRuleEngine(ctx, db, emailEntity.To, input)
	emailEntity.Subject = RunRuleEngine(ctx, db, emailEntity.Subject, input)
	emailEntity.Sender = RunRuleEngine(ctx, db, emailEntity.Sender, input)

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

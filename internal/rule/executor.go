package rule

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/mailgun/mailgun-go/v4"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/ruler"
)

func execute(ctx context.Context, db *sqlx.DB, expression string, input map[string]string) error {
	log.Println("executing executor for ", expression)
	entityID := ruler.FetchEntityID(expression)
	e, err := entity.Retrieve(ctx, entityID, db)
	if err != nil {
		return err
	}
	allFields, err := e.AllFields()
	if err != nil {
		return err
	}

	itemID := ruler.FetchItemID(expression)
	itemFieldsMap := map[string]interface{}{}
	if itemID != "" {
		i, err := item.Retrieve(ctx, itemID, db)
		if err != nil {
			return err
		}
		itemFieldsMap = i.Fields()
	}

	params := populateNamedFields(allFields, itemFieldsMap)

	switch e.Category {
	case entity.CategoryAPI:
	case entity.CategoryEmail:
		emailEntity, err := entity.MakeEmailEntity(params)
		if err != nil {
			log.Println("err ", err)
		}
		body := RunRuleEngine(ctx, db, emailEntity.Body, input)
		log.Println("body ---> ", body)
		to := RunRuleEngine(ctx, db, emailEntity.To, input)
		log.Println("to ---> ", to)
		subject := RunRuleEngine(ctx, db, emailEntity.Subject, input)
		log.Println("subject ---> ", subject)
		sender := RunRuleEngine(ctx, db, emailEntity.Sender, input)
		log.Println("sender ---> ", sender)
	}

	return err
}

func sendSimpleMessage(ctx context.Context, domain, apiKey string) (string, error) {
	mg := mailgun.NewMailgun(domain, apiKey)
	m := mg.NewMessage(
		"Excited User <mailgun@YOUR_DOMAIN_NAME>",
		"Hello",
		"Testing some Mailgun awesomeness!",
		"YOU@YOUR_DOMAIN_NAME",
	)
	_, id, err := mg.Send(ctx, m)
	return id, err
}

func populateNamedFields(entityFields []entity.Field, itemFields map[string]interface{}) map[string]interface{} {
	params := map[string]interface{}{}
	for _, field := range entityFields {
		if val, ok := itemFields[field.Key]; ok {
			params[field.Name] = val
		} else {
			params[field.Name] = field.Value
		}
	}
	return params
}

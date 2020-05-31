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
	itemID := ruler.FetchItemID(expression)
	e, err := entity.Retrieve(ctx, entityID, db)
	if err != nil {
		return err
	}
	fields, err := e.Fields()
	if err != nil {
		return err
	}

	i, err := item.Retrieve(ctx, itemID, db)
	if err != nil {
		return err
	}
	log.Println("i ---> ", i)
	log.Println("fields ---> ", fields)
	return err
}

func SendSimpleMessage(ctx context.Context, domain, apiKey string) (string, error) {
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

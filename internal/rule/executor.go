package rule

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/ruler"
)

func execute(ctx context.Context, db *sqlx.DB, expression string, input map[string]string) error {
	log.Println("executing executor for ", expression)
	action := ruler.ActionExpression(expression, input)
	e, err := entity.Retrieve(ctx, action.EntityID, db)
	if err != nil {
		return err
	}

	entityFields, err := e.AllFields()
	if err != nil {
		return err
	}

	switch e.Category {
	case entity.CategoryAPI:
		executeHook(ctx, db, entityFields)
	case entity.CategoryEmail:
		executeEmail(ctx, db, input, entityFields)
	case entity.CategoryData:
		executeData(ctx, db, input, entityFields, action)

	}

	return err
}

func namedFieldsMap(entityFields []entity.Field) map[string]interface{} {
	params := map[string]interface{}{}
	for _, field := range entityFields {
		params[field.Name] = field.Value
	}
	return params
}

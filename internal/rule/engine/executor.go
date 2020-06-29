package engine

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func execute(ctx context.Context, db *sqlx.DB, n node.Node) error {
	log.Println("execute ActorID ---> ", n.ActorID)
	e, err := entity.Retrieve(ctx, n.ActorID, db)

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
		executeEmail(ctx, db, entityFields, n)
	case entity.CategoryData:
		executeData(ctx, db, entityFields, n)

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

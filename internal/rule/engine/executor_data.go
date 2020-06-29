package engine

import (
	"context"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func executeData(ctx context.Context, db *sqlx.DB, entityFields []entity.Field, n node.Node) error {

	log.Println("variables === ", n.Variables)
	log.Printf("entityFields11 === %+v", entityFields)

	entityFields, err := fillItemFieldValues(ctx, db, entityFields, n.ActualsMap()[n.ActorID])
	if err != nil {
		return err
	}
	log.Printf("entityFields22 === %+v", entityFields)

	ni := item.NewItem{
		AccountID: n.AccountID,
		EntityID:  n.ActorID,
	}
	ni.Fields = evaluateFieldValues(ctx, db, entityFields, n)

	switch n.Type {
	case node.Push:
		i, err := item.Create(ctx, db, ni, time.Now())
		log.Printf("iiiii %v", i)
		log.Println("err", err)
	case node.Modify:
		actualItemID := n.ActualsMap()[n.ActorID]
		item.UpdateFields(ctx, db, actualItemID, ni.Fields)
		i, err := item.Retrieve(ctx, actualItemID, db)
		log.Printf("iiiii %v", i)
		log.Println("err", err)
	}

	return nil
}

func evaluateFieldValues(ctx context.Context, db *sqlx.DB, entityFields []entity.Field, n node.Node) map[string]interface{} {
	evaluatedItemFields := map[string]interface{}{}
	for _, field := range entityFields {
		switch field.DataType {
		case entity.TypeString:
			valuatedValue := RunRenderer(ctx, db, field.Value.(string), n.VariablesMap())
			evaluatedItemFields[field.Key] = valuatedValue
		default:
			evaluatedItemFields[field.Key] = field.Value
		}

	}
	return evaluatedItemFields
}

func fillItemFieldValues(ctx context.Context, db *sqlx.DB, entityFields []entity.Field, itemIDs ...string) ([]entity.Field, error) {
	for _, itemID := range itemIDs {
		if itemID != "" {
			i, err := item.Retrieve(ctx, itemID, db)
			if err != nil {
				return nil, err
			}
			entityFields = entity.FillFieldValues(entityFields, i.Fields())
		}
	}

	return entityFields, nil
}

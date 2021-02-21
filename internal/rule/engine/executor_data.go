package engine

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func executeData(ctx context.Context, db *sqlx.DB, n node.Node) error {
	entityFields, err := mergeActualsWithActor(ctx, db, n.AccountID, n.ActorID, n.ActualsItemID())
	if err != nil {
		return err
	}

	ni := item.NewItem{
		ID:        uuid.New().String(),
		AccountID: n.AccountID,
		EntityID:  n.ActorID,
	}
	ni.Fields = evaluateFieldValues(ctx, db, n.AccountID, entityFields, n.VariablesMap())

	log.Printf("ni %+v ", ni)

	switch n.Type {
	case node.Push:
		_, err = item.Create(ctx, db, ni, time.Now())
	case node.Modify:
		actualItemID := n.ActualsMap()[n.ActorID]
		err = item.UpdateFields(ctx, db, n.ActorID, actualItemID, ni.Fields)
		if err != nil {
			return err
		}
		_, err = item.Retrieve(ctx, n.ActorID, actualItemID, db)
	}

	return err
}

func evaluateFieldValues(ctx context.Context, db *sqlx.DB, accountID string, entityFields []entity.Field, vars map[string]interface{}) map[string]interface{} {
	evaluatedItemFields := map[string]interface{}{}
	for _, field := range entityFields {
		switch field.DataType {
		case entity.TypeString:
			if field.Value != nil {
				valuatedValue := RunExpRenderer(ctx, db, accountID, field.Value.(string), vars)
				evaluatedItemFields[field.Key] = valuatedValue
			}
		case entity.TypeReference:
			evaluatedItemFields[field.Key] = vars[field.RefID] // what happens if the vars has more than one item
		default:
			evaluatedItemFields[field.Key] = field.Value
		}

	}
	return evaluatedItemFields
}

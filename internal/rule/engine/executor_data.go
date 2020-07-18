package engine

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func executeData(ctx context.Context, db *sqlx.DB, n node.Node) error {
	entityFields, err := fields(ctx, db, n.ActorID)
	if err != nil {
		return err
	}
	entityFields, err = fillItemFieldValues(ctx, db, entityFields, n.ActualsMap()[n.ActorID])
	if err != nil {
		return err
	}

	ni := item.NewItem{
		AccountID: n.AccountID,
		EntityID:  n.ActorID,
	}
	ni.Fields = evaluateFieldValues(ctx, db, entityFields, n)

	switch n.Type {
	case node.Push:
		_, err = item.Create(ctx, db, ni, time.Now())
	case node.Modify:
		actualItemID := n.ActualsMap()[n.ActorID]
		item.UpdateFields(ctx, db, actualItemID, ni.Fields)
		_, err = item.Retrieve(ctx, actualItemID, db)
	}

	return err
}

func evaluateFieldValues(ctx context.Context, db *sqlx.DB, entityFields []entity.Field, n node.Node) map[string]interface{} {
	evaluatedItemFields := map[string]interface{}{}
	for _, field := range entityFields {
		switch field.DataType {
		case entity.TypeString:
			valuatedValue := RunExpRenderer(ctx, db, field.Value.(string), n.VariablesMap())
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

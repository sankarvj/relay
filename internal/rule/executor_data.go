package rule

import (
	"context"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/ruler"
)

func executeData(ctx context.Context, db *sqlx.DB, input map[string]string, entityFields []entity.Field, action ruler.Action) error {
	var err error
	if action.ItemID != "" {
		entityFields, err = fillItemFieldValues(ctx, db, entityFields, action.ItemID, action.SecItemID)
		if err != nil {
			return err
		}
	}

	log.Println("input === ", input)
	log.Printf("entityFields === %+v", entityFields)

	ni := item.NewItem{}
	ni.Fields = evaluateFieldValues(ctx, db, input, entityFields)
	if action.Behaviour == ruler.Update {
		err = item.UpdateFields(ctx, db, action.ItemID, ni.Fields)
		i, _ := item.Retrieve(ctx, action.ItemID, db)
		log.Printf("item -- %+v", i)
	} else if action.Behaviour == ruler.Create {
		i, err1 := item.Create(ctx, db, action.EntityID, ni, time.Now())
		err = err1
		log.Printf("item -- %+v", i)
	} else if action.Behaviour == ruler.Retrive {
		log.Println("TODO: Retrive")
	}

	log.Println("err--? ", err)

	if err != nil {
		return err
	}
	return nil
}

func evaluateFieldValues(ctx context.Context, db *sqlx.DB, input map[string]string, entityFields []entity.Field) map[string]interface{} {
	evaluatedItemFields := map[string]interface{}{}
	for _, field := range entityFields {
		switch field.DataType {
		case entity.TypeString:
			valuatedValue := RunRuleEngine(ctx, db, field.Value.(string), input)
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

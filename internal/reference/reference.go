package reference

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/rule/engine"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

/***
In this file we map the fields with the reference field value for the specific item
**/

func UpdateReferenceFields(ctx context.Context, fields []entity.Field, items []*item.ViewModelItem, db *sqlx.DB) {
	referenceFields := make(map[string]*entity.Field, 0)
	referenceIds := make(map[string][]interface{}, 0)

	for i := 0; i < len(fields); i++ {
		if fields[i].IsReference() || fields[i].IsPipe() {
			referenceIds[fields[i].Key] = []interface{}{}
			referenceFields[fields[i].Key] = &fields[i]
		}
	}

	for _, item := range items {
		for key, vals := range item.Fields {
			if refIds, ok := referenceIds[key]; ok {
				referenceIds[key] = append(refIds, vals.([]interface{})...)
			}
		}
	}

	for _, f := range referenceFields {
		if f.DomType == entity.DomSelect {
			updateChoicesForFieldUnits(ctx, db, f)
		} else if f.DomType == entity.DomAutoSelect {
			updateChoicesWithExpression(ctx, db, f, items)
		} else if f.DomType == entity.DomPipeline || f.DomType == entity.DomPlayBook {
			updateChoicesForPipeLine(ctx, db, f)
		} else {
			updateChoicesForOtherSelectDom(ctx, db, f, referenceIds)
		}
	}
}

//TODO: Is it efficient? As of now for field unit reference we need to query n+1 time
func updateChoicesForFieldUnits(ctx context.Context, db *sqlx.DB, f *entity.Field) {
	refItems, err := item.EntityItems(ctx, f.RefID, db)
	if err != nil {
		log.Println("error on retriving reference items for field unit entity. continuing... ", err)
	}

	for _, refItem := range refItems {
		f.Choices = append(f.Choices, entity.Choice{
			ID:           refItem.ID,
			DisplayValue: refItem.Fields()[f.DisplayGex()],
		})
	}
}

func updateChoicesWithExpression(ctx context.Context, db *sqlx.DB, f *entity.Field, items []*item.ViewModelItem) {
	choiceExpressions := f.Choices //store choice expressions and empty the choices
	f.Choices = make([]entity.Choice, 0)
	updateChoicesForFieldUnits(ctx, db, f)

	for i := 0; i < len(items); i++ {
		if len(items[i].Fields[f.Key].([]interface{})) > 0 { // Don't set auto if the value exist already
			continue
		}
		for _, choice := range choiceExpressions {
			result := engine.RunExpEvaluator(ctx, db, nil, choice.Expression, items[i].Fields)
			if result {
				items[i].Fields[f.Key] = []interface{}{choice.ID}
			}
		}
	}
}

func updateChoicesForPipeLine(ctx context.Context, db *sqlx.DB, f *entity.Field) {
	nodes, err := node.Stages(ctx, f.RefID, db)
	if err != nil {
		log.Println("error on retriving reference items for field unit entity. continuing... ", err)
	}
	for _, node := range nodes {
		f.Choices = append(f.Choices, entity.Choice{
			ID:           node.ID,
			DisplayValue: node.Name,
		})
	}
}

func updateChoicesForOtherSelectDom(ctx context.Context, db *sqlx.DB, f *entity.Field, referenceIds map[string][]interface{}) {
	refItems, err := item.BulkRetrieve(ctx, f.RefID, removeDuplicateValues(referenceIds[f.Key]), db)
	if err != nil {
		log.Println("error on retriving reference items for selected items. continuing... ", err)
	}

	if f.Choices == nil {
		f.Choices = make([]entity.Choice, 0)
	}

	for _, refItem := range refItems {
		f.Choices = append(f.Choices, entity.Choice{
			ID:           refItem.ID,
			DisplayValue: refItem.Fields()[f.DisplayGex()],
		})
	}
}

func removeDuplicateValues(intSlice []interface{}) []interface{} {
	keys := make(map[interface{}]bool)
	list := []interface{}{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

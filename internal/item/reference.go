package item

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

/***
In this file we map the fields with the reference field value for the specific item
**/

func UpdateReferenceFields(ctx context.Context, fields []*entity.Field, items []ViewModelItem, db *sqlx.DB) {
	referenceFields := make(map[string]*entity.Field, 0)
	referenceIds := make(map[string][]interface{}, 0)

	tmpFields := fields[:0]
	for _, f := range fields {
		if f.IsNotApplicable() { // remove not appicable fields from the view
			continue
		}

		if f.IsReference() || f.IsPipe() {
			referenceIds[f.Key] = []interface{}{}
			referenceFields[f.Key] = f
		}
		tmpFields = append(tmpFields, f)
	}
	fields = tmpFields

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
		} else if f.DomType == entity.DomPipeline || f.DomType == entity.DomPlayBook {
			updateChoicesForPipeLine(ctx, db, f)
		} else {
			updateChoicesForOtherSelectDom(ctx, db, f, referenceIds)
		}
	}
}

//TODO: Is it efficient? As of now for field unit reference we need to query n+1 time
func updateChoicesForFieldUnits(ctx context.Context, db *sqlx.DB, f *entity.Field) {
	refItems, err := entityItems(ctx, f.RefID, db)
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
	refItems, err := BulkRetrieve(ctx, f.RefID, removeDuplicateValues(referenceIds[f.Key]), db)
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

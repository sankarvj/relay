package reference

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/rule/engine"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

/***
In this file we map the fields with the reference field value for the specific item
**/

func UpdateReferenceFields(ctx context.Context, accountID string, fields []entity.Field, items []*item.ViewModelItem, db *sqlx.DB) {
	referenceFields := make(map[string]*entity.Field, 0)
	referenceIds := make(map[string][]interface{}, 0)

	for i := 0; i < len(fields); i++ {
		if fields[i].IsReference() {
			referenceIds[fields[i].Key] = []interface{}{}
			referenceFields[fields[i].Key] = &fields[i]
		}
	}

	// Fetch the selected item values only
	for _, item := range items {
		for key, vals := range item.Fields {
			if refIds, ok := referenceIds[key]; ok {
				if vals == nil {
					vals = []interface{}{}
				}
				referenceIds[key] = append(refIds, vals.([]interface{})...)
			}
		}
	}

	for _, f := range referenceFields {
		if f.DomType == entity.DomSelect {
			updateChoicesForFieldUnits(ctx, db, accountID, f)
		} else if f.DomType == entity.DomAutoSelect {
			updateChoicesWithExpression(ctx, db, accountID, f, items)
		} else { //auto complete will come here.
			updateChoices(ctx, db, accountID, f, referenceIds[f.Key])
		}
	}

}

//TODO: Is it efficient? As of now for field unit reference we need to query n+1 time
func updateChoicesForFieldUnits(ctx context.Context, db *sqlx.DB, accountID string, f *entity.Field) {
	refItems, err := item.EntityItems(ctx, f.RefID, db)
	if err != nil {
		log.Println("error on retriving reference items for field unit entity. continuing... ", err)
	}
	choicesMaker(f, refItems)
}

func updateChoicesWithExpression(ctx context.Context, db *sqlx.DB, accountID string, f *entity.Field, items []*item.ViewModelItem) {
	choiceExpressions := f.Choices //store choice expressions and empty the choices
	f.Choices = make([]entity.Choice, 0)
	updateChoicesForFieldUnits(ctx, db, accountID, f)

	for i := 0; i < len(items); i++ {
		if len(items[i].Fields[f.Key].([]interface{})) > 0 { // Don't execute the choice expressions and set the value if it is already set. for more details go to README
			continue
		}
		for _, choice := range choiceExpressions {
			result := engine.RunExpEvaluator(ctx, db, nil, accountID, choice.Expression, items[i].Fields)
			if result {
				items[i].Fields[f.Key] = []interface{}{choice.ID}
			}
		}
	}
}

//updateChoices simply update single value to the choice if that itemID is populated already.
//updateChoices won't pull all the choices available to that reference entity in the list view.
//updateChoices bulk get all the references for the particular item and updates the choices once for each reference field
//updateChoices should work differently in the detail use-case
func updateChoices(ctx context.Context, db *sqlx.DB, accountID string, f *entity.Field, referenceIds []interface{}) {

	e, err := entity.Retrieve(ctx, accountID, f.RefID, db)
	if err != nil {
		log.Println("error on retriving entity when updatingChoices. continuing... ", err)
		return
	}

	if e.Category == entity.CategoryNode { // temp flow handler
		nodes, err := node.BulkRetrieve(ctx, referenceIds, db)
		if err != nil {
			log.Println("error on retriving reference items for field unit entity. continuing... ", err)
			return
		}
		choicesMakerNode(f, nodes)
	} else if e.Category == entity.CategoryFlow { // temp flow handler
		flows, err := flow.BulkRetrieve(ctx, accountID, removeDuplicateValues(referenceIds), db)
		if err != nil {
			log.Println("error on retriving flows when updatingChoices. continuing... ", err)
			return
		}
		choicesMakerFlow(f, flows)
	} else {
		refItems, err := item.BulkRetrieve(ctx, f.RefID, removeDuplicateValues(referenceIds), db)
		if err != nil {
			log.Println("error on retriving reference items when updatingChoices. continuing... ", err)
			return
		}
		choicesMaker(f, refItems)
	}

}

func choicesMaker(f *entity.Field, items []item.Item) {
	if f.Choices == nil {
		f.Choices = make([]entity.Choice, 0)
	}

	for _, item := range items {
		f.Choices = append(f.Choices, entity.Choice{
			ID:           item.ID,
			Verb:         item.Fields()[f.Verb()],
			DisplayValue: item.Fields()[f.DisplayGex()],
		})
	}
}

func choicesMakerFlow(f *entity.Field, flows []flow.Flow) {
	if f.Choices == nil {
		f.Choices = make([]entity.Choice, 0)
	}

	for _, flow := range flows {
		f.Choices = append(f.Choices, entity.Choice{
			ID:           flow.ID,
			DisplayValue: flow.Name,
		})
	}
}

func choicesMakerNode(f *entity.Field, nodes []node.Node) {
	if f.Choices == nil {
		f.Choices = make([]entity.Choice, 0)
	}

	for _, node := range nodes {
		f.Choices = append(f.Choices, entity.Choice{
			ID:           node.ID,
			DisplayValue: node.Name,
		})
	}
}

//UpdateChoicesWrapper updates only the choices for reference fields
func UpdateChoicesWrapper(ctx context.Context, db *sqlx.DB, accountID string, valueAddedFields []entity.Field) {
	for i := 0; i < len(valueAddedFields); i++ {
		if valueAddedFields[i].IsReference() {
			updateChoices(ctx, db, accountID, &valueAddedFields[i], valueAddedFields[i].Value.([]interface{}))
		}
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

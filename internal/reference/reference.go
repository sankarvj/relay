package reference

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/rule/engine"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

const (
	ActionFilter = "filter"
	ActionView   = "view"
	ActionSet    = "set"
)

const (
	ByFlow      = "flow_id"
	ByDoNothing = "donothing"
	ByShow      = "show"
	ByHide      = "hide"
)

/***
In this file we map the fields with the reference field value for the specific item
If the items count is more than 1, we will just update the value for the current selected item only.
If the items count is equal to 1, we will populate all the available choices if the referenced entity is child unit.
**/

func UpdateReferenceFields(ctx context.Context, accountID, entityID string, fields []entity.Field, items []item.Item, srcMap map[string]interface{}, db *sqlx.DB, eng *engine.Engine) {

	refIds := populateExistingItemIds(items, fields)
	// 1.populate referenceIDS in the map format <fieldKey: []ExistingItemIds{} + []BaseItemIds{}>
	// 2.based on the value populated update the choices
	for i := 0; i < len(fields); i++ {
		f := &fields[i]
		ids := refIds[f.Key]
		if srcItemID, ok := srcMap[f.RefID]; ok { // if the parent item exists please add it to the item
			if ids == nil {
				ids = []interface{}{}
			}
			ids = append(ids, srcItemID)
		}
		updateChoices(ctx, db, accountID, entityID, f, ids, eng)
	}

	//dependent logic during list/retrive/edit. this logic will not get executed in create
	//the values evaluted using this logic should be attached to each item
	for _, item := range items {
		for i := 0; i < len(fields); i++ {
			f := &fields[i]
			if f.IsDependent() {
				_, pv := parentField(fields, f.Dependent.ParentKey, item.Fields())

				for k, exp := range f.Dependent.Expressions {
					result := eng.RunExpEvaluator(ctx, db, nil, accountID, exp, item.Fields())

					if result {
						action := f.Dependent.Actions[k] // take the corresponding action

						if pv != "" {
							switch action {
							case fmt.Sprintf("{{{%s.%s}}}", ActionSet, ByDoNothing):
							case fmt.Sprintf("{{{%s.%s}}}", ActionView, ByShow):
							case fmt.Sprintf("{{{%s.%s}}}", ActionView, ByHide):
							case fmt.Sprintf("{{{%s.%s}}}", ActionFilter, ByFlow):
								nodes, err := node.Stages(ctx, pv, db)
								if err != nil {
									log.Printf("***> unexpected error occurred. when retriving reference nodes for field unit entity. continuing... error: %v\n", err)
									return
								}
								choicesMaker(f, pv, nodeChoices(nodes))
							case fmt.Sprintf("{{{%s.%s}}}", ActionSet, ByFlow):
							}

						}
						//if one result is evaluated stop ther for that field in the item
						break
					}
				}

			}
		}
	}
}

//TODO not efficient. Put it in a map???
func populateExistingItemIds(items []item.Item, fields []entity.Field) map[string][]interface{} {
	referenceIds := make(map[string][]interface{}, 0)
	for _, i := range items {
		if (i == item.Item{}) {
			continue
		}
		for _, f := range fields {
			if f.IsReference() && f.RefID != "" && i.Fields()[f.Key] != nil {
				if _, ok := referenceIds[f.Key]; ok { // if the parent item exists please add it to the item
					referenceIds[f.Key] = append(referenceIds[f.Key], i.Fields()[f.Key].([]interface{})...)
				} else {
					referenceIds[f.Key] = i.Fields()[f.Key].([]interface{})
				}
			}
		}
	}
	return referenceIds
}

// // TODO not understandable. please update the readme
// func evaluateChoices(ctx context.Context, db *sqlx.DB, accountID string, f *entity.Field, items []item.ViewModelItem, eng *engine.Engine) []interface{} {
// 	refIDs := make([]interface{}, 0)
// 	for i := 0; i < len(items); i++ {
// 		if len(items[i].Fields[f.Key].([]interface{})) > 0 { // Don't execute the choice expressions and set the value if it is already set. for more details go to README
// 			continue
// 		}
// 		for _, choice := range f.Choices {
// 			//items[i].Fields[f.Key] = []interface{}{choice.ID}
// 			refIDs = append(refIDs, choice.ID)
// 			//result := eng.RunExpEvaluator(ctx, db, nil, accountID, choice.Expression, items[i].Fields)
// 		}
// 	}
// 	return refIDs
// }

//updateChoices simply update single value to the choice if that itemID if populated already.
//updateChoices won't pull all the choices available to that reference entity in the list view.
//updateChoices bulk get all the references for the particular item and updates the choices once for each reference field
//updateChoices should work differently in the detail use-case
func updateChoices(ctx context.Context, db *sqlx.DB, accountID, entityID string, f *entity.Field, refIDs []interface{}, eng *engine.Engine) {

	if f.IsReference() && f.RefID != "" && !f.IsNotApplicable() {
		e, err := entity.Retrieve(ctx, accountID, f.RefID, db)
		if err != nil {
			log.Printf("***> unexpected error occurred when retriving entity inside updating choices error: %v.\n continuing...", err)
			return
		}
		if e.Category == entity.CategoryChildUnit {
			refItems, err := item.EntityItems(ctx, e.ID, db)
			if err != nil {
				log.Printf("***> unexpected error occurred when retriving reference items for field unit inside updating choices error: %v.\n continuing...", err)
			}
			choicesMaker(f, "", itemChoices(*f, refItems, e.WhoFields()))
		} else if e.Category == entity.CategoryEmail {
			refItems, err := item.EntityItems(ctx, e.ID, db)
			if err != nil {
				log.Printf("***> unexpected error occurred when retriving reference items for email entity inside updating choices error: %v.\n continuing...", err)
			}
			choicesMaker(f, "", itemChoices(*f, refItems, e.WhoFields()))
		} else { // useful for auto-complete while viewing

			refItems, err := item.BulkRetrieve(ctx, e.ID, removeDuplicateValues(refIDs), db)
			if err != nil && err != item.ErrItemsEmpty {
				log.Printf("***> unexpected error occurred when retriving reference items inside updating choices error: %v.\n continuing...", err)
				return
			}

			choicesMaker(f, "", itemChoices(*f, refItems, e.WhoFields()))
		}

		//RETHINK
		if e.Category == entity.CategoryFlow { //  flow handler
			var flows []flow.Flow
			var err error
			if len(refIDs) == 0 { // create or edit case.
				flows, err = flow.List(ctx, []string{entityID}, flow.FlowModeAll, flow.FlowTypeUnknown, db)
			} else { // view case
				flows, err = flow.BulkRetrieve(ctx, accountID, removeDuplicateValues(refIDs), db) // though the name bulk retrive is misleading this fetches only the flow which is selected
			}
			if err != nil {
				log.Printf("***> unexpected error occurred when retriving flows inside updating choices error: %v.\n continuing...", err)
				return
			}
			choicesMaker(f, "", flowChoices(flows))
		}
	}

}

func parentField(fields []entity.Field, key string, vals map[string]interface{}) (*entity.Field, string) {
	for _, f := range fields {
		if f.Key == key {
			return &f, evaluatedParentVal(&f, vals[key])
		}
	}
	return nil, ""
}

func evaluatedParentVal(pf *entity.Field, pv interface{}) string {
	var evaluatedDependentValue string
	if pf.IsList() || pf.IsReference() {
		if pv != nil && len(pv.([]interface{})) > 0 && pv.([]interface{})[0] != nil {
			evaluatedDependentValue = pv.([]interface{})[0].(string)
		}
	} else {
		if pv != nil {
			evaluatedDependentValue = pv.(string)
		}
	}
	return evaluatedDependentValue
}

func choicesMaker(f *entity.Field, parentID string, choicers []Choicer) {
	if f.Choices == nil {
		f.Choices = make([]entity.Choice, 0)
	}

	mapOfChoices := choicesMap(f)

	for _, choicer := range choicers {
		if choice, ok := mapOfChoices[f.RefID]; ok {
			if isIdNotExistAlready(choice.ParentIDs, parentID) {
				choice.ParentIDs = append(choice.ParentIDs, parentID)
			}
		} else {

			f.Choices = append(f.Choices, entity.Choice{
				ID:           choicer.ID,
				ParentIDs:    []string{parentID},
				DisplayValue: choicer.Name,
				Value:        choicer.Value,
				Verb:         util.ConvertIntfToStr(choicer.Verb),
				Avatar:       choicer.Avatar,
			})
		}
	}
}

//ChoicesBluePrint populate the choices of the template fields and also sets the one by evaluting the parent which is creating it
func ChoicesBluePrint(f *entity.Field, sourceEntity entity.Entity) {
	sourceFields := sourceEntity.FieldsIgnoreError()

	//Adding parents existing items to the child choices
	for _, sf := range sourceFields {
		if sf.RefID == f.RefID {
			f.Choices = append(f.Choices, entity.Choice{
				ID:           fmt.Sprintf("{{%s.%s}}", sourceEntity.ID, sf.Key),
				DisplayValue: strings.Title(strings.ToLower(fmt.Sprintf("%s's existing %s", sourceEntity.DisplayName, f.DisplayName))),
				BaseChoice:   true,
			})

			if f.IsNode() {
				f.Value = []interface{}{fmt.Sprintf("{{%s.%s}}", sourceEntity.ID, sf.Key)}
				f.SetMeta("config")
			}

			break
		}
	}

	//Set the parent ID if the child has implicit dependency
	if sourceEntity.ID == f.RefID {
		f.Value = []interface{}{fmt.Sprintf("{{%s.%s}}", sourceEntity.ID, "id")}
		f.SetMeta("config")
		f.Choices = append(f.Choices, entity.Choice{
			ID:           fmt.Sprintf("{{%s.%s}}", sourceEntity.ID, "id"),
			DisplayValue: strings.Title(strings.ToLower(fmt.Sprintf("existing %s ", sourceEntity.DisplayName))),
			Default:      true,
		})
	}

}

//UpdateChoicesWrapper updates only the choices for reference fields
func UpdateChoicesWrapper(ctx context.Context, db *sqlx.DB, accountID, entityID string, valueAddedFields []entity.Field, eng *engine.Engine) {
	for i := 0; i < len(valueAddedFields); i++ {
		if valueAddedFields[i].IsReference() {
			var refIds []interface{}
			if valueAddedFields[i].Value != nil {
				refIds = valueAddedFields[i].Value.([]interface{})
			}

			updateChoices(ctx, db, accountID, entityID, &valueAddedFields[i], refIds, eng)
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

func choicesMap(f *entity.Field) map[string]entity.Choice {
	choicesMap := make(map[string]entity.Choice, 0)
	for _, choice := range f.Choices {
		choicesMap[choice.ID] = choice
	}
	return choicesMap
}

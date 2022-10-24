package reference

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
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

func UpdateReferenceFields(ctx context.Context, accountID, entityID string, fields []entity.Field, items []item.Item, srcMap map[string]interface{}, db *sqlx.DB, sdb *database.SecDB, eng *engine.Engine) {

	//populate base entity only in the blue print case
	var be entity.Entity
	if len(srcMap) > 0 && len(items) > 0 && items[0].State == item.StateBluePrint {
		keys := make([]string, 0, len(srcMap))
		for k := range srcMap {
			keys = append(keys, k)
		}
		be, _ = entity.Retrieve(ctx, accountID, keys[0], db, sdb)
	}

	refIds := populateExistingItemIds(items, fields)
	// 1.populate referenceIDS in the map format <fieldKey: []ExistingItemIds{} + []BaseItemIds{}>
	// 2.based on the value populated update the choices
	for i := 0; i < len(fields); i++ {
		f := &fields[i]
		ids := refIds[f.Key]
		if srcItemID, ok := srcMap[f.RefID]; ok && srcItemID != "" { // if the parent item exists please add it to the item. why? we are silently associating the parent item with the child
			if ids == nil {
				ids = []interface{}{}
			}
			ids = append(ids, srcItemID)
		}
		updateChoices(ctx, db, sdb, accountID, entityID, f, ids, eng)

		//updating base choices for blue print case
		if len(items) > 0 && items[0].State == item.StateBluePrint {
			if f.IsReference() {
				updateBPChoices(&fields[i], &be)
			} else if f.IsNode() {
				updateBPChoices(&fields[i], &be)
			}
		}
	}

	//dependent logic during list/retrive/edit. this logic should not get executed in create or blueprint/webform create
	//the values evaluted using this logic should be attached to each item
	for _, i := range items {

		if i.ID == "" || i.State == item.StateWebForm || i.State == item.StateBluePrint { // skip create/bp
			continue
		}

		for index := 0; index < len(fields); index++ {
			f := &fields[index]
			if f.IsDependent() {
				_, pv := parentField(fields, f.Dependent.ParentKey, i.Fields())
				for k, exp := range f.Dependent.Expressions {
					log.Printf("******> debug internal.reference called `RunExpEvaluator` for dependents evalution key: %+v exp: %+v ", k, exp)
					result := eng.RunExpEvaluator(ctx, db, nil, accountID, exp, i.Fields())
					if result {
						action := f.Dependent.Actions[k] // take the corresponding action

						if pv != "" {
							switch action {
							case fmt.Sprintf("{{{%s.%s}}}", ActionSet, ByDoNothing):
							case fmt.Sprintf("{{{%s.%s}}}", ActionView, ByShow):
							case fmt.Sprintf("{{{%s.%s}}}", ActionView, ByHide):
							case fmt.Sprintf("{{{%s.%s}}}", ActionFilter, ByFlow):
								nodes, err := node.Stages(ctx, accountID, []string{pv}, "", db)
								if err != nil {
									log.Printf("***> unexpected error occurred. when retriving reference nodes for field unit entity. continuing... error: %v\n", err)
									return
								}
								ChoicesMaker(f, pv, nodeChoices(nodes))
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
		if (i == item.Item{} || i.State == item.StateWebForm) {
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

//updateChoices simply update single value to the choice if that itemID if populated already.
//updateChoices won't pull all the choices available to that reference entity in the list view.
//updateChoices bulk get all the references for the particular item and updates the choices once for each reference field
//updateChoices should work differently in the detail use-case
func updateChoices(ctx context.Context, db *sqlx.DB, sdb *database.SecDB, accountID, entityID string, f *entity.Field, refIDs []interface{}, eng *engine.Engine) {

	if f.IsReference() && f.RefID != "" && !f.IsNotApplicable() {
		e, err := entity.Retrieve(ctx, accountID, f.RefID, db, sdb)
		if err != nil {
			log.Printf("***> unexpected error occurred when retriving entity inside updating choices error: %v.\n continuing...", err)
			return
		}
		if e.Category == entity.CategoryFlow { // temp flow handler
			// fi is the entityID here
			flows, err := flow.SearchByKey(ctx, accountID, entityID, "", db)
			if err != nil {
				log.Printf("***> unexpected error occurred when retriving reference items for flows inside updating choices error: %v.\n continuing...", err)
			}
			ChoicesMaker(f, "", flowChoices(flows))
		} else if e.Category == entity.CategoryNode { // temp flow handler

		} else if e.Category == entity.CategoryChildUnit {
			refItems, err := item.EntityItems(ctx, accountID, e.ID, db)
			if err != nil {
				log.Printf("***> unexpected error occurred when retriving reference items for field unit inside updating choices error: %v.\n continuing...", err)
			}
			ChoicesMaker(f, "", ItemChoices(f, refItems, e.WhoFields()))
		} else if e.Category == entity.CategoryEmail {
			refItems, err := item.EntityItems(ctx, accountID, e.ID, db)
			if err != nil {
				log.Printf("***> unexpected error occurred when retriving reference items for email entity inside updating choices error: %v.\n continuing...", err)
			}
			ChoicesMaker(f, "", ItemChoices(f, refItems, e.WhoFields()))
		} else { // useful for auto-complete while viewing
			refItems, err := item.BulkRetrieve(ctx, e.ID, removeDuplicateValues(refIDs), db)
			if err != nil && err != item.ErrItemsEmpty {
				log.Printf("***> unexpected error occurred when retriving reference items inside updating choices error: %v.\n continuing...", err)
				return
			}

			ChoicesMaker(f, "", ItemChoices(f, refItems, e.WhoFields()))
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

func ChoicesMaker(f *entity.Field, parentID string, choicers []Choicer) {
	if f.Choices == nil {
		f.Choices = make([]entity.Choice, 0)
	}

	mapOfChoices := f.ChoicesMap()

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
				Color:        util.ConvertIntfToStr(choicer.Color),
			})
		}
	}
}

//updateBPChoices populate the choices of the template fields and also sets the one by evaluting the parent which is creating it
func updateBPChoices(f *entity.Field, sourceEntity *entity.Entity) {

	if sourceEntity == nil {
		return
	}

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
func UpdateChoicesWrapper(ctx context.Context, db *sqlx.DB, sdb *database.SecDB, accountID, entityID string, valueAddedFields []entity.Field, eng *engine.Engine) {
	for i := 0; i < len(valueAddedFields); i++ {
		if valueAddedFields[i].IsReference() {
			var refIds []interface{}
			if valueAddedFields[i].Value != nil {
				refIds = valueAddedFields[i].Value.([]interface{})
			}
			updateChoices(ctx, db, sdb, accountID, entityID, &valueAddedFields[i], refIds, eng)
		}
	}
}

func removeDuplicateValues(intSlice []interface{}) []interface{} {
	keys := make(map[interface{}]bool)
	list := []interface{}{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			if _, err := uuid.Parse(entry.(string)); err == nil { //invalidating expression ids {{something.id}}
				keys[entry] = true
				list = append(list, entry)
			}
		}
	}
	return list
}

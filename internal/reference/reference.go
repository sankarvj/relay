package reference

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/rule/engine"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

/***
In this file we map the fields with the reference field value for the specific item
If the items count is more than 1, we will just update the value for the current selected item only.
If the items count is equal to 1, we will populate all the available choices if the referenced entity is child unit.
**/

//Check dependent struct in the fields.go to understand it fully
const (
	ActionShow       = "show"
	ActionHide       = "hide"
	ActionLoadStages = "stages"
)

func UpdateReferenceFields(ctx context.Context, accountID, entityID string, fields []entity.Field, items []item.ViewModelItem, srcMap map[string]interface{}, db *sqlx.DB, eng *engine.Engine) {
	referenceFields := make(map[string]*entity.Field, 0)
	referenceIds := make(map[string][]interface{}, 0)

	for i := 0; i < len(fields); i++ {
		if fields[i].IsReference() && fields[i].RefID != "" {
			if srcItemID, ok := srcMap[fields[i].RefID]; ok { // if the parent item exists please add it to the item
				referenceIds[fields[i].Key] = []interface{}{srcItemID}
			} else {
				referenceIds[fields[i].Key] = []interface{}{}
			}
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
		if f.DomType == entity.DomAutoSelect {
			refIDs := evaluateChoices(ctx, db, accountID, f, items, eng)
			referenceIds[f.Key] = append(referenceIds[f.Key], refIDs...)
		}

		if f.Dependent != nil {
			evaluateDependentValue(ctx, db, accountID, f, items, eng)
		}

		updateChoices(ctx, db, accountID, entityID, f, referenceIds[f.Key])
	}

}

// TODO not understandable. please update the readme
func evaluateChoices(ctx context.Context, db *sqlx.DB, accountID string, f *entity.Field, items []item.ViewModelItem, eng *engine.Engine) []interface{} {
	refIDs := make([]interface{}, 0)
	for i := 0; i < len(items); i++ {
		if len(items[i].Fields[f.Key].([]interface{})) > 0 { // Don't execute the choice expressions and set the value if it is already set. for more details go to README
			continue
		}
		for _, choice := range f.Choices {
			result := eng.RunExpEvaluator(ctx, db, nil, accountID, choice.Expression, items[i].Fields)
			if result {
				items[i].Fields[f.Key] = []interface{}{choice.ID}
				refIDs = append(refIDs, choice.ID)
			}
		}
	}
	return refIDs
}

func evaluateDependentValue(ctx context.Context, db *sqlx.DB, accountID string, f *entity.Field, items []item.ViewModelItem, eng *engine.Engine) {
	for i := 0; i < len(items); i++ {
		parentField := items[i].Fields[f.Dependent.ParentKey]
		if parentField == nil || len(parentField.([]interface{})) == 0 || parentField.([]interface{})[0] == nil {
			continue
		}
		//TODO what happens if more than one value exists???
		f.Dependent.EvalutedValue = parentField.([]interface{})[0].(string)
		//f.Dependent.EvalutedExpression = eng.RunExpEvaluator(ctx, db, nil, accountID, f.Dependent.Expression, items[i].Fields)

	}
}

//updateChoices simply update single value to the choice if that itemID is populated already.
//updateChoices won't pull all the choices available to that reference entity in the list view.
//updateChoices bulk get all the references for the particular item and updates the choices once for each reference field
//updateChoices should work differently in the detail use-case
func updateChoices(ctx context.Context, db *sqlx.DB, accountID, entityID string, f *entity.Field, choiceIds []interface{}) {

	if f.IsNotApplicable() {
		return
	}

	e, err := entity.Retrieve(ctx, accountID, f.RefID, db)
	if err != nil {
		log.Println("error on retriving entity when updatingChoices. continuing... ", err)
		return
	}

	if e.Category == entity.CategoryNode { //  node handler
		var nodes []node.Node
		var err error
		if len(choiceIds) == 0 && f.Dependent.EvalutedValue != "" { // only for edit case.
			nodes, err = node.Stages(ctx, f.Dependent.EvalutedValue, db)
		} else { // view case
			nodes, err = node.Stages(ctx, f.Dependent.EvalutedValue, db)
			//nodes, err = node.BulkRetrieve(ctx, choiceIds, db)
		}
		if err != nil {
			log.Println("error on retriving reference nodes for field unit entity. continuing... ", err)
			return
		}
		choicesMakerNode(f, nodes)
	} else if e.Category == entity.CategoryFlow { //  flow handler
		var flows []flow.Flow
		var err error
		if len(choiceIds) == 0 { // create or edit case.
			flows, err = flow.List(ctx, []string{entityID}, flow.FlowModeAll, db)
		} else { // view case
			flows, err = flow.BulkRetrieve(ctx, accountID, removeDuplicateValues(choiceIds), db) // though the name bulk retrive is misleading this fetches only the flow which is selected
		}
		if err != nil {
			log.Println("error on retriving flows when updatingChoices. continuing... ", err)
			return
		}
		choicesMakerFlow(f, flows)
	} else if e.Category == entity.CategoryChildUnit { // select. with pre-populated drop-down choices
		refFields := e.FieldsIgnoreError()
		refItems, err := item.EntityItems(ctx, e.ID, db)
		//TODO bring dependent field logic here.
		if f.Dependent != nil {

		}

		if err != nil {
			log.Println("error on retriving reference items for field unit entity. continuing... ", err)
		}
		choicesMaker(f, refItems, refFields)
	} else { //auto-complete or multi-select. The UI must provide the search for these fields
		if len(choiceIds) == 0 { // fetch with some limit
			// I hope we can get some items for multi-select with limit here.
			f.Choices = make([]entity.Choice, 0)
		} else {
			refFields := e.FieldsIgnoreError()
			refItems, err := item.BulkRetrieve(ctx, e.ID, removeDuplicateValues(choiceIds), db)
			if err != nil {
				log.Println("error on retriving reference items when updatingChoices. continuing... ", err)
				return
			}
			choicesMaker(f, refItems, refFields)
		}
	}
}

func choicesMaker(f *entity.Field, refItems []item.Item, refFields []entity.Field) {

	var verbFieldKey string
	for _, refField := range refFields {
		if refField.Name == entity.Verb {
			verbFieldKey = refField.Key
		}
	}

	f.Choices = make([]entity.Choice, 0)
	for _, refItem := range refItems {
		f.Choices = append(f.Choices, entity.Choice{
			ID:           refItem.ID,
			Verb:         refItem.Fields()[verbFieldKey],
			DisplayValue: refItem.Fields()[f.DisplayGex()],
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

//ChoicesBluePrint populate the choices of the template fields and also sets the one by evaluting the parent which is creating it
func ChoicesBluePrint(f *entity.Field, sourceEntity entity.Entity) {
	sourceFields := sourceEntity.FieldsIgnoreError()

	//Adding parents existing items to the child choices
	for _, sf := range sourceFields {
		if sf.RefID == f.RefID {

			f.Choices = append(f.Choices, entity.Choice{
				ID:           fmt.Sprintf("{{%s.%s}}", sourceEntity.ID, sf.Key),
				Verb:         "",
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
			Verb:         "",
			DisplayValue: strings.Title(strings.ToLower(fmt.Sprintf("existing %s ", sourceEntity.DisplayName))),
			Default:      true,
		})
	}

}

//UpdateChoicesWrapper updates only the choices for reference fields
func UpdateChoicesWrapper(ctx context.Context, db *sqlx.DB, accountID, entityID string, valueAddedFields []entity.Field) {
	for i := 0; i < len(valueAddedFields); i++ {
		if valueAddedFields[i].IsReference() {
			var refIds []interface{}
			if valueAddedFields[i].Value != nil {
				refIds = valueAddedFields[i].Value.([]interface{})
			}

			updateChoices(ctx, db, accountID, entityID, &valueAddedFields[i], refIds)
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

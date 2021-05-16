package job

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/connection"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/integration/calendar"
	"gitlab.com/vjsideprojects/relay/internal/integration/email"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/ruler"
	"gitlab.com/vjsideprojects/relay/internal/reference"
	"gitlab.com/vjsideprojects/relay/internal/relationship"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
)

//func's in this package should not throw errors. It should handle errors by re-queue/dl-queue

func EventItemUpdated(accountID, entityID, itemID string, newFields, oldFields map[string]interface{}, db *sqlx.DB) {
	ctx := context.Background()
	it, err := item.Retrieve(ctx, entityID, itemID, db)
	if err != nil {
		log.Println("error while retriving item on job", err)
		return
	}
	if it.State == item.StateBluePrint {
		return
	}

	e, err := entity.Retrieve(ctx, accountID, entityID, db)
	if err != nil {
		log.Println("error while retriving entity on job", err)
		return
	}
	validateWorkflows(ctx, db, e, itemID, oldFields, newFields)
	addConnection(ctx, db, accountID, map[string]string{}, entityID, itemID, e.ValueAdd(newFields), e.ValueAdd(oldFields))
}

func EventItemCreated(accountID, entityID string, ni item.NewItem, db *sqlx.DB) {
	ctx := context.Background()
	if ni.State == item.StateBluePrint {
		return
	}

	e, err := entity.Retrieve(ctx, accountID, entityID, db)
	if err != nil {
		log.Println("error while retriving entity on job", err)
		return
	}
	valueAddedFields := e.ValueAdd(ni.Fields)
	//validateWorkflows(ctx, db, entityID, itemID, oldFields, newFields)
	addConnection(ctx, db, accountID, ni.Source, entityID, ni.ID, valueAddedFields, nil)
	reference.UpdateChoicesWrapper(ctx, db, accountID, valueAddedFields)

	//integrations
	switch e.Category {
	case entity.CategoryEmail:
		err = email.SendMail(ctx, accountID, e.ID, ni.ID, valueAddedFields, db)
	case entity.CategoryMeeting:
		err = calendar.CreateCalendarEvent(ctx, accountID, e.ID, ni.ID, valueAddedFields, db)
	}
	if err != nil {
		log.Println("error while performing the job", err)
	}
}

func validateWorkflows(ctx context.Context, db *sqlx.DB, e entity.Entity, itemID string, oldFields, newFields map[string]interface{}) {
	// log.Println("entityID...", entityID)
	// log.Println("itemID...", itemID)
	// log.Println("oldFields...", oldFields)
	// log.Println("newFields...", newFields)

	//workflows
	flows, err := flow.List(context.Background(), []string{e.ID}, -1, db)
	if err != nil {
		log.Println("There is an error while selecting flows...", err)
	}
	dirtyFlows := flow.DirtyFlows(context.Background(), flows, item.Diff(oldFields, newFields))
	if len(dirtyFlows) > 0 {
		log.Print("Tick...\nTick...\nTick...\nTick...\nTick...\nTick...\n The flow trigger has been started")
		errs := flow.Trigger(context.Background(), db, nil, itemID, dirtyFlows)
		if len(errs) > 0 {
			log.Println("There is an error while triggering flows...", errs)
		}
	}

	//pipelines
	for _, fi := range e.FieldsIgnoreError() {
		if fi.IsNode() {
			flowID := newFields[fi.Dependent.ParentKey].([]interface{})[0].(string)
			nodeID := newFields[fi.Key].([]interface{})[0].(string)
			flow.DirectTrigger(context.Background(), db, nil, e.AccountID, flowID, nodeID, e.ID, itemID)
		}
	}

}

// It connects the implicit relationships which as inferred by the field
func addConnection(ctx context.Context, db *sqlx.DB, accountID string, base map[string]string, entityID, itemID string, newFields, oldFields []entity.Field) {
	createEvent := oldFields == nil
	newValueAddedFieldsMap := entity.KeyedFieldsObjMap(newFields)
	oldValueAddedFieldsMap := entity.KeyedFieldsObjMap(oldFields)
	relationships, err := relationship.Relationships(ctx, db, accountID, entityID)
	if err != nil {
		log.Println("There is an error while querying relationships...", err)
		return
	}

	for _, r := range relationships {
		//Explicit connection. Happens when adding the item from inside the base element
		//The user can only delete that association and he couldn't update it
		//Hence no reference exists between the two entities implicitly.
		if r.FieldID == relationship.FieldAssociationKey && createEvent {
			if baseItemID, ok := base[r.DstEntityID]; ok {
				err = connection.Associate(ctx, db, accountID, r.RelationshipID, itemID, baseItemID)
				if err != nil {
					log.Println("Explicit association failed ", err)
				}
			}
		} else {
			//Implicit connection with straight reference. When create a deal with contact as its reference field
			if f, ok := newValueAddedFieldsMap[r.FieldID]; ok {
				if f.IsFlow() || f.IsNode() {
					log.Println("Handle Flow/Node here")
				} else if f.ValidRefField() && r.DstEntityID == f.RefID {
					if of, ok := oldValueAddedFieldsMap[r.FieldID]; ok { //handle update case
						f.Value = compare(ctx, db, accountID, r.RelationshipID, f, of) //update the f.Value with only the updated value
					}
					for _, dstItemID := range f.RefValues() {
						err := connection.Associate(ctx, db, accountID, r.RelationshipID, itemID, dstItemID.(string))
						if err != nil {
							log.Println("Implicit connection with straight reference failed", err)
							return
						}
					}
				}
			} else { //Implicit connection with reverse reference. When creating the contact inside a deal base
				log.Println("Implicit connection with reverse reference")
				if baseItemID, ok := base[r.DstEntityID]; ok && createEvent { //This won't happen during the update
					err = connection.Associate(ctx, db, accountID, r.RelationshipID, itemID, baseItemID)
					baseItem, err := item.Retrieve(ctx, r.DstEntityID, baseItemID, db)
					if err != nil {
						log.Println("Implicit connection with reverse reference failed ", err)
					}
					itemFieldsMap := baseItem.Fields()
					log.Println("BF itemFieldsMap ", itemFieldsMap)
					if vals, ok := itemFieldsMap[r.FieldID]; ok { // little complex
						exisitingVals := vals.([]interface{})
						exisitingVals = append(exisitingVals, itemID)
						itemFieldsMap[r.FieldID] = exisitingVals
						log.Println("AF itemFieldsMap ", itemFieldsMap)
						_, err = item.UpdateFields(ctx, db, r.DstEntityID, baseItemID, itemFieldsMap)
						if err != nil {
							log.Println("Implicit connection with reverse reference failed ", err)
						}
					}
				}
			}
		}
	}
}

func compare(ctx context.Context, db *sqlx.DB, accountID, relationshipID string, f, of entity.Field) []interface{} {
	if ruler.Compare(f.Value, of.Value) { // handle delete alone here
		deletedItems, newItems := item.CompareItems(f.Value.([]interface{}), of.Value.([]interface{}))
		for _, deletedItem := range deletedItems {
			err := connection.Delete(ctx, db, relationshipID, deletedItem.(string))
			if err != nil {
				log.Println("error while deleting connection", err)
			}
		}
		return newItems
	}
	return []interface{}{}
}

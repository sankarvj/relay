package job

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/connection"
	"gitlab.com/vjsideprojects/relay/internal/email"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/ruler"
	"gitlab.com/vjsideprojects/relay/internal/reference"
	"gitlab.com/vjsideprojects/relay/internal/relationship"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
)

//func's in this package should not throw errors. It should handle errors by re-queue/dl-queue

func EventItemUpdated(accountID, entityID, itemID string, newFields, oldFields map[string]interface{}, db *sqlx.DB) {
	ctx := context.Background()

	e, err := entity.Retrieve(ctx, accountID, entityID, db)
	if err != nil {
		log.Println("error while retriving entity on job", err)
		return
	}
	newValueAddedFields := entity.ValueAddFields(e.FieldsIgnoreError(), newFields)
	oldValueAddedFields := entity.ValueAddFields(e.FieldsIgnoreError(), oldFields)

	validateWorkflows(ctx, db, entityID, itemID, oldFields, newFields)
	addConnection(ctx, db, accountID, map[string]string{}, entityID, itemID, oldValueAddedFields, newValueAddedFields)
}

func EventItemCreated(accountID, entityID string, ni item.NewItem, db *sqlx.DB) {
	ctx := context.Background()

	e, err := entity.Retrieve(ctx, accountID, entityID, db)
	if err != nil {
		log.Println("error while retriving entity on job", err)
		return
	}
	valueAddedFields := entity.ValueAddFields(e.FieldsIgnoreError(), ni.Fields)
	//validateWorkflows(ctx, db, entityID, itemID, oldFields, newFields)
	addConnection(ctx, db, accountID, ni.Source, entityID, ni.ID, valueAddedFields, nil)

	reference.UpdateChoicesWrapper(ctx, db, accountID, valueAddedFields)

	switch e.Category {
	case entity.CategoryEmail:
		err = email.SendMail(ctx, accountID, e.ID, ni.ID, valueAddedFields, db)
	}
	if err != nil {
		log.Println("error while performing the job", err)
	}

}

func validateWorkflows(ctx context.Context, db *sqlx.DB, entityID, itemID string, oldFields, newFields map[string]interface{}) {
	// log.Println("entityID...", entityID)
	// log.Println("itemID...", itemID)
	// log.Println("oldFields...", oldFields)
	// log.Println("newFields...", newFields)
	flows, err := flow.List(context.Background(), []string{entityID}, -1, db)
	if err != nil {
		log.Println("There is an error while selecting flows...", err)
	}
	dirtyFlows := flow.DirtyFlows(context.Background(), flows, oldFields, newFields)

	log.Printf("This update triggers %d flows", len(dirtyFlows))
	if len(dirtyFlows) > 0 {
		log.Print("Tick...\nTick...\nTick...\nTick...\nTick...\nTick...\n")

		log.Println("The flow trigger has been started")
		errs := flow.Trigger(context.Background(), db, nil, itemID, dirtyFlows)

		if len(errs) > 0 {
			log.Println("There is an error while triggering flows...", errs)
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
		if r.FieldID == relationship.FieldAssociationKey && createEvent { //Explicit association. This won't happen during the update
			if baseItemID, ok := base[r.DstEntityID]; ok {
				err = connection.Associate(ctx, db, accountID, r.RelationshipID, itemID, baseItemID)
				if err != nil {
					log.Println("Explicit association failed ", err)
				}
			}
		} else { //Implicit association
			if f, ok := newValueAddedFieldsMap[r.FieldID]; ok { //Implicit association with straight reference. When create a deal with contact as its reference field
				if f.IsFlow() || f.IsNode() {
					log.Println("Handle Flow/Node here")
				} else if f.ValidRefField() && r.DstEntityID == f.RefID {
					if of, ok := oldValueAddedFieldsMap[r.FieldID]; ok { //handle update case
						f.Value = compare(ctx, db, accountID, r.RelationshipID, f, of) //update the f.Value with only the updated value
					}
					for _, dstItemID := range f.RefValues() {
						c := connection.Connection{
							AccountID:      accountID,
							RelationshipID: r.RelationshipID,
							SrcItemID:      itemID,
							DstItemID:      dstItemID.(string),
						}

						_, err := connection.Create(ctx, db, c)
						if err != nil {
							log.Println("Implicit association with straight reference failed", err)
							return
						}
					}
				}
			} else { //Implicit association with reverse reference. When creating the contact inside a deal base
				log.Println("Implicit association with reverse reference")
				if baseItemID, ok := base[r.DstEntityID]; ok && createEvent { //This won't happen during the update
					err = connection.Associate(ctx, db, accountID, r.RelationshipID, itemID, baseItemID)
					baseItem, err := item.Retrieve(ctx, r.DstEntityID, baseItemID, db)
					if err != nil {
						log.Println("Implicit association with reverse reference failed ", err)
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
							log.Println("Implicit association with reverse reference failed ", err)
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

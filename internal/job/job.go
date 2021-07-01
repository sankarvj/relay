package job

import (
	"context"
	"log"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/connection"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/integration/calendar"
	"gitlab.com/vjsideprojects/relay/internal/integration/email"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/ruler"
	"gitlab.com/vjsideprojects/relay/internal/reference"
	"gitlab.com/vjsideprojects/relay/internal/relationship"
	"gitlab.com/vjsideprojects/relay/internal/rule/engine"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
)

//func's in this package should not throw errors. It should handle errors by re-queue/dl-queue
type Job struct {
}

func (j *Job) EventItemUpdated(accountID, entityID, itemID string, newFields, oldFields map[string]interface{}, db *sqlx.DB, rp *redis.Pool) {
	ctx := context.Background()
	e, err := entity.Retrieve(ctx, accountID, entityID, db)
	if err != nil {
		log.Println("error while retriving entity on job", err)
		return
	}
	it, err := item.Retrieve(ctx, entityID, itemID, db)
	if err != nil {
		log.Println("error while retriving item on job", err)
		return
	}
	if it.State == item.StateBluePrint {
		return
	}
	j.validateWorkflows(e, itemID, oldFields, newFields, db, rp)
	j.AddConnection(accountID, map[string]string{}, entityID, itemID, e.ValueAdd(newFields), e.ValueAdd(oldFields), db)

	//insertion in to redis graph DB
	insertInToRedisGraph(accountID, entityID, it.ID, e.ValueAdd(newFields), rp)
}

func (j *Job) EventItemCreated(accountID, entityID string, it item.Item, source map[string]string, db *sqlx.DB, rp *redis.Pool) {
	ctx := context.Background()
	if it.State == item.StateBluePrint {
		return
	}

	e, err := entity.Retrieve(ctx, accountID, entityID, db)
	if err != nil {
		log.Println("error while retriving entity on job", err)
		return
	}
	valueAddedFields := e.ValueAdd(it.Fields())
	reference.UpdateChoicesWrapper(ctx, db, accountID, valueAddedFields)
	//j.validateWorkflows(db, entityID, itemID, oldFields, newFields)
	j.AddConnection(accountID, source, entityID, it.ID, valueAddedFields, nil, db)

	//integrations
	switch e.Category {
	case entity.CategoryEmail:
		err = email.SendMail(ctx, accountID, e.ID, it.ID, valueAddedFields, db)
	case entity.CategoryMeeting:
		err = calendar.CreateCalendarEvent(ctx, accountID, e.ID, it.ID, valueAddedFields, db)
	}
	if err != nil {
		log.Println("error while performing the job", err)
	}

	//insertion in to redis graph DB
	insertInToRedisGraph(accountID, entityID, it.ID, valueAddedFields, rp)
}

func insertInToRedisGraph(accountID, entityID, itemID string, valueAddedFields []entity.Field, rp *redis.Pool) {
	gpbNode := graphdb.BuildGNode(accountID, entityID, false).MakeBaseGNode(itemID, makeGraphFields(valueAddedFields))
	err := graphdb.UpsertNode(rp, gpbNode)
	if err != nil {
		log.Println("error while performing the rDB insertion job", err)
	}
}

func makeGraphFields(fields []entity.Field) []graphdb.Field {
	gFields := make([]graphdb.Field, len(fields))
	for i, f := range fields {
		gFields[i] = *makeGraphField(&f)
	}
	return gFields
}

func makeGraphField(f *entity.Field) *graphdb.Field {
	if f == nil {
		return nil
	}

	return &graphdb.Field{
		Key:      f.Key,
		Value:    f.Value,
		DataType: graphdb.DType(f.DataType),
		RefID:    f.RefID,
		Field:    makeGraphField(f.Field),
	}
}

func (j *Job) validateWorkflows(e entity.Entity, itemID string, oldFields, newFields map[string]interface{}, db *sqlx.DB, rp *redis.Pool) {
	// log.Println("entityID...", entityID)
	// log.Println("itemID...", itemID)
	// log.Println("oldFields...", oldFields)
	// log.Println("newFields...", newFields)
	eng := engine.Engine{
		Job: j,
	}
	//workflows
	flows, err := flow.List(context.Background(), []string{e.ID}, -1, db)
	if err != nil {
		log.Println("There is an error while selecting flows...", err)
	}
	dirtyFields := item.Diff(oldFields, newFields)
	dirtyFlows := flow.DirtyFlows(context.Background(), flows, dirtyFields)

	log.Println("dirtyFlows --", dirtyFlows)
	if len(dirtyFlows) > 0 {
		log.Print("Tick...\nTick...\nTick...\nTick...\nTick...\nTick...\n The flow trigger has been started")

		errs := flow.Trigger(context.Background(), db, nil, itemID, dirtyFlows, eng)
		if len(errs) > 0 {
			log.Println("There is an error while triggering flows...", errs)
		}
	}

	//pipelines -  not a generic way. the way we use dependent is muddy
	for _, fi := range e.FieldsIgnoreError() {
		if dirtyField, ok := dirtyFields[fi.Key]; ok && fi.IsNode() {
			flowID := newFields[fi.Dependent.ParentKey].([]interface{})[0].(string)
			nodeID := dirtyField.([]interface{})[0].(string)
			err = flow.DirectTrigger(context.Background(), db, nil, e.AccountID, flowID, nodeID, e.ID, itemID, eng)
			if err != nil {
				log.Println("There is an error while triggering flows...", err)
			}
		}
	}

}

// It connects the implicit relationships which as inferred by the field
func (j Job) AddConnection(accountID string, base map[string]string, entityID, itemID string, newFields, oldFields []entity.Field, db *sqlx.DB) {
	ctx := context.Background()
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

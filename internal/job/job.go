package job

import (
	"context"
	"fmt"
	"log"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
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
		log.Println("unexpected error occurred when retriving entity inside job. error:", err)
		return
	}

	err = j.actOnWorkflows(ctx, e, itemID, oldFields, newFields, db, rp)
	if err != nil {
		log.Println(err)
		return
	}
	err = j.actOnConnections(accountID, map[string]string{}, entityID, itemID, e.ValueAdd(newFields), e.ValueAdd(oldFields), db)
	if err != nil {
		log.Println(err)
		return
	}
	//insertion in to redis graph DB
	err = j.actOnRedisGraph(accountID, entityID, itemID, e.ValueAdd(newFields), "", "", rp)
	if err != nil {
		log.Println(err)
		return
	}
}

func (j *Job) EventItemCreated(accountID, entityID, itemID string, source map[string]string, db *sqlx.DB, rp *redis.Pool) {
	ctx := context.Background()

	e, err := entity.Retrieve(ctx, accountID, entityID, db)
	if err != nil {
		log.Println("unexpected error occurred when retriving entity on job. error:", err)
		return
	}
	it, err := item.Retrieve(ctx, entityID, itemID, db)
	if err != nil {
		log.Println("unexpected error occurred while retriving item on job. error:", err)
		return
	}

	valueAddedFields := e.ValueAdd(it.Fields())
	reference.UpdateChoicesWrapper(ctx, db, accountID, entityID, valueAddedFields, NewJabEngine())

	err = j.actOnWorkflows(ctx, e, itemID, nil, it.Fields(), db, rp)
	if err != nil {
		log.Println(err)
		return
	}

	j.actOnConnections(accountID, source, entityID, itemID, valueAddedFields, nil, db)
	if err != nil {
		log.Println(err)
		return
	}

	//integrations
	err = actOnIntegrations(ctx, accountID, e, it, db)
	if err != nil {
		log.Println(err)
		return
	}

	//insertion in to redis graph DB
	for baseEntityID, baseItemID := range source {
		err = j.actOnRedisGraph(accountID, entityID, itemID, valueAddedFields, baseEntityID, baseItemID, rp)
		if err != nil {
			log.Println(err)
			return
		}
	}

}

func (j *Job) eventCreated(ctx context.Context, baseEntityID, baseItemID string, evItem item.Item, db *sqlx.DB, rp *redis.Pool) error {
	ae, err := entity.Retrieve(ctx, evItem.AccountID, evItem.EntityID, db)
	if err != nil {
		return err
	}
	avalueAddedFields := ae.ValueAdd(evItem.Fields())
	avalueAddedFields = append(avalueAddedFields)
	gpbNode := graphdb.BuildGNode(ae.AccountID, ae.ID, false).MakeBaseGNode(evItem.ID, makeGraphFields(avalueAddedFields)).ParentEdge(baseEntityID, baseItemID)
	err = graphdb.UpsertNode(rp, gpbNode)
	if err != nil {
		return errors.Wrap(err, "error: redisGrpah insertion job")
	}
	return nil
}

func (j *Job) actOnRedisGraph(accountID, entityID, itemID string, valueAddedFields []entity.Field, baseEntityID, baseItemID string, rp *redis.Pool) error {
	gpbNode := graphdb.BuildGNode(accountID, entityID, false).MakeBaseGNode(itemID, makeGraphFields(valueAddedFields))
	if baseEntityID != "" && baseItemID != "" {
		gpbNode = gpbNode.ParentEdge(baseEntityID, baseItemID)
	}
	err := graphdb.UpsertNode(rp, gpbNode)
	if err != nil {
		return errors.Wrap(err, "error: redisGrpah insertion job")
	}
	return nil
}

func (j *Job) actOnWorkflows(ctx context.Context, e entity.Entity, itemID string, oldFields, newFields map[string]interface{}, db *sqlx.DB, rp *redis.Pool) error {
	eng := engine.Engine{
		Job: j,
	}
	flowType := flow.FlowTypeEventUpdate //eventUpdate
	if oldFields == nil {                //eventCreate
		flowType = flow.FlowTypeEventCreate
	}

	//workflows - eventCreate/eventUpdate
	flows, err := flow.List(ctx, []string{e.ID}, flow.FlowModeWorkFlow, flowType, db)
	if err != nil {
		return err
	}

	var errs []error
	dirtyFields := item.Diff(oldFields, newFields)
	if len(flows) > 0 {
		switch flowType {
		case flow.FlowTypeEventUpdate:
			dirtyFlows := flow.DirtyFlows(ctx, flows, dirtyFields)
			if len(dirtyFlows) > 0 {
				errs = flow.Trigger(ctx, db, rp, itemID, dirtyFlows, eng)
			}
		case flow.FlowTypeEventCreate:
			errs = flow.Trigger(ctx, db, rp, itemID, flows, eng)
		}
	}

	err = actOnPipelines(ctx, eng, e, itemID, dirtyFields, newFields, db, rp)
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		for i, err := range errs {
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("actOnWorkflows error index: %d error msg: %+v", i, err))
			}
		}

	}
	return nil
}

//actOnPipelines -  not a generic way. the way we use dependent is muddy
func actOnPipelines(ctx context.Context, eng engine.Engine, e entity.Entity, itemID string, dirtyFields map[string]interface{}, newFields map[string]interface{}, db *sqlx.DB, rp *redis.Pool) error {
	for _, fi := range e.FieldsIgnoreError() {
		if dirtyField, ok := dirtyFields[fi.Key]; ok && fi.IsNode() && len(dirtyField.([]interface{})) > 0 && fi.Dependent != nil {
			flowID := newFields[fi.Dependent.ParentKey].([]interface{})[0].(string)
			nodeID := dirtyField.([]interface{})[0].(string)
			err := flow.DirectTrigger(ctx, db, rp, e.AccountID, flowID, nodeID, e.ID, itemID, eng)
			if err != nil {
				return errors.Wrap(err, "error: acting on pipelines")
			}
		}
	}
	return nil
}

// It connects the implicit relationships which as inferred by the field
func (j Job) actOnConnections(accountID string, base map[string]string, entityID, itemID string, newFields, oldFields []entity.Field, db *sqlx.DB) error {
	ctx := context.Background()
	createEvent := oldFields == nil
	newValueAddedFieldsMap := entity.KeyedFieldsObjMap(newFields)
	oldValueAddedFieldsMap := entity.KeyedFieldsObjMap(oldFields)
	relationships, err := relationship.Relationships(ctx, db, accountID, entityID)
	if err != nil {
		return errors.Wrap(err, "error: querying relationships")
	}

	for _, r := range relationships {
		//Explicit connection happens when adding the item from inside the base element
		//The user can only delete that association and he couldn't update it becasue there is no
		//reference exists between the two entities implicitly.
		if r.FieldID == relationship.FieldAssociationKey && createEvent {
			if baseItemID, ok := base[r.DstEntityID]; ok {
				err = connection.Associate(ctx, db, accountID, r.RelationshipID, itemID, baseItemID)
				if err != nil {
					return errors.Wrap(err, "error: querying association")
				}
			}
		} else {
			//Implicit connection with straight reference. When create a deal with contact as its reference field
			if f, ok := newValueAddedFieldsMap[r.FieldID]; ok {
				if f.IsFlow() || f.IsNode() {
					log.Println("internal.job handle flow/node here")
				} else if f.ValidRefField() && r.DstEntityID == f.RefID {
					if of, ok := oldValueAddedFieldsMap[r.FieldID]; ok { //handle update case
						f.Value = compare(ctx, db, accountID, r.RelationshipID, f, of) //update the f.Value with only the updated value
					}
					for _, dstItemID := range f.RefValues() {
						err := connection.Associate(ctx, db, accountID, r.RelationshipID, itemID, dstItemID.(string))
						if err != nil {
							return errors.Wrap(err, "error: implicit connection with straight reference failed")
						}
					}
				}
			} else { //Implicit connection with reverse reference. When creating the contact inside a deal base
				//log.Println("internal.job implicit connection with reverse reference handled")
				if baseItemID, ok := base[r.DstEntityID]; ok && createEvent { //This won't happen during the update
					err = connection.Associate(ctx, db, accountID, r.RelationshipID, itemID, baseItemID)
					baseItem, err := item.Retrieve(ctx, r.DstEntityID, baseItemID, db)
					if err != nil {
						return errors.Wrap(err, "error: implicit connection with reverse reference failed")
					}
					itemFieldsMap := baseItem.Fields()
					log.Println("internal.job BF itemFieldsMap ", itemFieldsMap)
					if vals, ok := itemFieldsMap[r.FieldID]; ok { // little complex
						exisitingVals := vals.([]interface{})
						exisitingVals = append(exisitingVals, itemID)
						itemFieldsMap[r.FieldID] = exisitingVals
						log.Println("internal.job AF itemFieldsMap ", itemFieldsMap)
						_, err = item.UpdateFields(ctx, db, r.DstEntityID, baseItemID, itemFieldsMap)
						if err != nil {
							return errors.Wrap(err, "error: implicit connection with reverse reference failed")
						}
					}
				}
			}
		}
	}
	return nil
}

func compare(ctx context.Context, db *sqlx.DB, accountID, relationshipID string, f, of entity.Field) []interface{} {
	if ruler.Compare(f.Value, of.Value) { // handle delete alone here
		deletedItems, newItems := item.CompareItems(f.Value.([]interface{}), of.Value.([]interface{}))
		for _, deletedItem := range deletedItems {
			err := connection.Delete(ctx, db, relationshipID, deletedItem.(string))
			if err != nil {
				log.Println("unexpected error occurred when deleting connection. error:", err)
			}
		}
		return newItems
	}
	return []interface{}{}
}

func actOnIntegrations(ctx context.Context, accountID string, e entity.Entity, it item.Item, db *sqlx.DB) error {
	valueAddedFields := e.ValueAdd(it.Fields())
	var err error
	switch e.Category {
	case entity.CategoryEmail:
		err = email.SendMail(ctx, accountID, e.ID, it.ID, valueAddedFields, db)
	case entity.CategoryMeeting:
		err = calendar.CreateCalendarEvent(ctx, accountID, e.TeamID, e.ID, it.ID, valueAddedFields, db)
	}
	return err
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

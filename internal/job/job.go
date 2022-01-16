package job

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/connection"
	conv "gitlab.com/vjsideprojects/relay/internal/conversation"
	"gitlab.com/vjsideprojects/relay/internal/discovery"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/integration/calendar"
	"gitlab.com/vjsideprojects/relay/internal/integration/email"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/notification"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/reference"
	"gitlab.com/vjsideprojects/relay/internal/relationship"
	"gitlab.com/vjsideprojects/relay/internal/rule/engine"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/user"
)

//func's in this package should not throw errors. It should handle errors by re-queue/dl-queue
type Job struct {
}

// events

func (j *Job) EventItemUpdated(accountID, entityID, itemID string, newFields, oldFields map[string]interface{}, db *sqlx.DB, rp *redis.Pool) {
	ctx := context.Background()
	e, err := entity.Retrieve(ctx, accountID, entityID, db)
	if err != nil {
		log.Println("EventItemUpdated: unexpected error occurred when retriving entity inside job. error:", err)
		return
	}

	valueAddedFields := e.ValueAdd(newFields)

	//workflows
	err = j.actOnWorkflows(ctx, e, itemID, oldFields, newFields, db, rp)
	if err != nil {
		log.Println("EventItemUpdated: unexpected error occurred on actOnWorkflows. error: ", err)
		return
	}

	//connections
	err = j.actOnConnections(accountID, map[string]string{}, entityID, itemID, valueAddedFields, e.ValueAdd(oldFields), db)
	if err != nil {
		log.Println("EventItemUpdated: unexpected error occurred on actOnConnections. error: ", err)
		return
	}

	//who
	err = j.actOnWho(accountID, entityID, itemID, valueAddedFields, rp)
	if err != nil {
		log.Println("EventItemUpdated: unexpected error occurred on actOnWho. error: ", err)
		return
	}

	//graph
	err = j.actOnRedisGraph(accountID, entityID, itemID, valueAddedFields, "", "", rp)
	if err != nil {
		log.Println("EventItemUpdated: unexpected error occurred on actOnRedisGraph. error: ", err)
		return
	}
}

func (j *Job) EventItemCreated(accountID, entityID, itemID string, source map[string]string, db *sqlx.DB, rp *redis.Pool) {
	ctx := context.Background()

	e, err := entity.Retrieve(ctx, accountID, entityID, db)
	if err != nil {
		log.Println("EventItemCreated: unexpected error occurred when retriving entity on job. error:", err)
		return
	}
	it, err := item.Retrieve(ctx, entityID, itemID, db)
	if err != nil {
		log.Println("EventItemCreated: unexpected error occurred while retriving item on job. error:", err)
		return
	}

	valueAddedFields := e.ValueAdd(it.Fields())
	reference.UpdateChoicesWrapper(ctx, db, accountID, entityID, valueAddedFields, NewJabEngine())

	//workflows
	err = j.actOnWorkflows(ctx, e, itemID, nil, it.Fields(), db, rp)
	if err != nil {
		log.Println("EventItemCreated: unexpected error occurred on actOnWorkflows. error: ", err)
		return
	}

	//connect
	err = j.actOnConnections(accountID, source, entityID, itemID, valueAddedFields, nil, db)
	if err != nil {
		log.Println("EventItemCreated: unexpected error occurred on actOnConnections. error: ", err)
		return
	}

	//integrations
	err = actOnIntegrations(ctx, accountID, e, it, valueAddedFields, db)
	if err != nil {
		log.Println("EventItemCreated: unexpected error occurred on actOnIntegrations. error: ", err)
		return
	}

	//who
	err = j.actOnWho(accountID, entityID, itemID, valueAddedFields, rp)
	if err != nil {
		log.Println("EventItemCreated: unexpected error occurred on actOnWho. error: ", err)
		return
	}

	//insertion in to redis graph DB
	if len(source) == 0 {
		err = j.actOnRedisGraph(accountID, entityID, itemID, valueAddedFields, "", "", rp)
		if err != nil {
			log.Println("EventItemCreated: unexpected error occurred on actOnRedisGraph. error: ", err)
			return
		}
	} else {
		for baseEntityID, baseItemID := range source {
			err = j.actOnRedisGraph(accountID, entityID, itemID, valueAddedFields, baseEntityID, baseItemID, rp)
			if err != nil {
				log.Println("EventItemCreated: unexpected error occurred on actOnRedisGraph. error: ", err)
				return
			}
		}
	}
}

func (j *Job) EventItemReminded(accountID, entityID, itemID string, db *sqlx.DB, rp *redis.Pool) {
	ctx := context.Background()

	e, err := entity.Retrieve(ctx, accountID, entityID, db)
	if err != nil {
		log.Println("EventItemReminded: unexpected error occurred when retriving entity on job. error:", err)
		return
	}
	it, err := item.Retrieve(ctx, entityID, itemID, db)
	if err != nil {
		log.Println("EventItemReminded: unexpected error occurred while retriving item on job. error:", err)
		return
	}

	valueAddedFields := e.ValueAdd(it.Fields())
	reference.UpdateChoicesWrapper(ctx, db, accountID, entityID, valueAddedFields, NewJabEngine())

	//save the notification to the notifications.
	err = notification.ItemUpdates(ctx, e.Name, accountID, e.TeamID, e.ID, it.ID, valueAddedFields, notification.TypeReminder, db)
	if err != nil {
		log.Println("EventItemReminded: unexpected error occurred on EventItemReminded. error: ", err)
	}
}

func (j *Job) EventItemDeleted(accountID, entityID, itemID string, db *sqlx.DB, rp *redis.Pool) {
	ctx := context.Background()

	e, err := entity.Retrieve(ctx, accountID, entityID, db)
	if err != nil {
		log.Println("EventItemDeleted: unexpected error occurred when retriving entity on job. error:", err)
		return
	}

	it, err := item.Retrieve(ctx, entityID, itemID, db)
	if err != nil {
		log.Println("EventItemDeleted: unexpected error occurred while retriving item on job. error:", err)
		return
	}

	err = destructOnIntegrations(ctx, accountID, e, it, db)
	if err != nil {
		log.Println("EventItemDeleted: unexpected error occurred on destructOnIntegrations. error: ", err)
		return
	}

	err = item.Delete(ctx, db, accountID, entityID, itemID)
	if err != nil {
		log.Println("EventItemDeleted: unexpected error occurred on delete main item. error: ", err)
		return
	}
}

func (j *Job) EventConvAdded(accountID, entityID, itemID, conversationID string, db *sqlx.DB) {
	ctx := context.Background()
	e, err := entity.Retrieve(ctx, accountID, entityID, db)
	if err != nil {
		log.Println("EventConvAdded: unexpected error occurred on retriving the entity on job. error:", err)
		return
	}

	var parentEmailEntityItem entity.EmailEntity
	_, err = entity.RetrieveUnmarshalledItem(ctx, accountID, entityID, itemID, &parentEmailEntityItem, db)
	if err != nil {
		log.Println("EventConvAdded: unexpected error occurred on retriving the parent entity on job. error:", err)
		return
	}

	cv, err := conv.Retrieve(ctx, accountID, conversationID, db)
	if err != nil {
		log.Println("EventConvAdded: unexpected error occurred while retriving item on job. error:", err)
		return
	}

	//TODO push to job
	replyTo := parentEmailEntityItem.MessageID
	valueAddedFields := e.ValueAdd(cv.PayloadMap())
	_, err = email.SendMail(ctx, accountID, entityID, itemID, valueAddedFields, replyTo, db)
	if err != nil {
		log.Println("Error while sending the mail - ", err)
	}
}

func (j *Job) EventUserInvited(usr user.User, db *sqlx.DB) {
	ctx := context.Background()
	err := notification.UserInvitation(ctx)
	if err != nil {
		log.Println("unexpected error occurred on EventUserInvited. error: ", err)
	}
}

func (j *Job) EventEmailReceived(db *sqlx.DB) {

}

//act ons

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
	log.Println("actOnWorkflows Kicked IN.......")
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

func actOnIntegrations(ctx context.Context, accountID string, e entity.Entity, it item.Item, valueAddedFields []entity.Field, db *sqlx.DB) error {
	var err error
	switch e.Category {
	case entity.CategoryEmail:
		if it.Name == nil || *it.Name != "received" { //super hacky :( Trying to avoid the sendmail action when saving the received mail
			msgID, err := email.SendMail(ctx, accountID, e.ID, it.ID, valueAddedFields, "", db)
			if err == nil {
				err = saveMsgID(ctx, accountID, e.ID, it.ID, *msgID, db)

			}
		}
	case entity.CategoryMeeting:
		err = calendar.CreateCalendarEvent(ctx, accountID, e.TeamID, e.ID, it.ID, valueAddedFields, db)
	}
	return err
}

func (j Job) actOnWho(accountID, entityID, itemID string, valueAddedFields []entity.Field, rp *redis.Pool) error {
	for _, f := range valueAddedFields {
		if f.Who == entity.WhoReminder && f.DataType == entity.TypeDateTime && f.Value != nil {
			when, err := util.ParseTime(f.Value.(string))
			if err != nil {
				return err
			}
			return (Listener{}).AddReminder(accountID, entityID, itemID, when, rp)
		}
	}
	return nil
}

func destructOnIntegrations(ctx context.Context, accountID string, e entity.Entity, it item.Item, db *sqlx.DB) error {
	var err error
	switch e.Category {
	case entity.CategoryEmailConfig:
		err = email.Destruct(ctx, accountID, e.ID, it.ID, db)
	case entity.CategoryCalendar:
		//calendar destruct yet to be implemented
	}
	return err
}

func saveMsgID(ctx context.Context, accountID, entityID, itemID, msgID string, db *sqlx.DB) error {
	ns := discovery.NewDiscovery{
		ID:        msgID,
		AccountID: accountID,
		EntityID:  entityID,
		ItemID:    itemID,
	}

	_, err := discovery.Create(ctx, db, ns, time.Now())
	if err != nil {
		return err
	}
	var emailItem entity.EmailEntity
	upFunc, err := entity.RetrieveUnmarshalledItem(ctx, accountID, entityID, itemID, emailItem, db)
	if err != nil {
		return err
	}
	emailItem.MessageID = msgID
	return upFunc(ctx, emailItem, db)
}

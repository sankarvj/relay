package job

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/account"
	"gitlab.com/vjsideprojects/relay/internal/connection"
	"gitlab.com/vjsideprojects/relay/internal/conversation"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/integration/calendar"
	"gitlab.com/vjsideprojects/relay/internal/integration/email"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/notification"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/stream"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/reference"
	"gitlab.com/vjsideprojects/relay/internal/relationship"
	"gitlab.com/vjsideprojects/relay/internal/rule/engine"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"gitlab.com/vjsideprojects/relay/internal/team"
	"gitlab.com/vjsideprojects/relay/internal/timeseries"
	"gitlab.com/vjsideprojects/relay/internal/user"
)

//func's in this package should not throw errors. It should handle errors by re-queue/dl-queue
type Job struct {
	baseItemIDs     []string
	baseEntityID    string
	DB              *sqlx.DB
	SDB             *database.SecDB
	FirebaseSDKPath string
}

func NewJob(db *sqlx.DB, sdb *database.SecDB, firebaseSDKPath string) *Job {
	return &Job{DB: db, SDB: sdb, FirebaseSDKPath: firebaseSDKPath}
}

func (j *Job) Post(msg *stream.Message) error {
	switch msg.Type {
	case stream.TypeItemCreate:
		return j.eventItemCreated(msg)
	case stream.TypeItemUpdate:
		return j.eventItemUpdated(msg)
	case stream.TypeItemDelete:
		return j.eventItemDeleted(msg)
	case stream.TypeItemRemind:
		return j.eventItemReminded(msg)
	case stream.TypeItemDelayed:
		return j.eventDelayExhausted(msg)
	case stream.TypeEmailConversationAdded:
		return j.eventEmailConvAdded(msg)
	case stream.TypeChatConversationAdded:
		return j.eventChatConvAdded(msg)
	case stream.TypeEventAdded:
		return j.eventEventAdded(msg)
	}
	return nil
}

// events
func (j *Job) eventItemCreated(m *stream.Message) error {
	log.Println("***>***> Reached EventItemCreated ***<***<")
	ctx := context.Background()

	e, err := entity.Retrieve(ctx, m.AccountID, m.EntityID, j.DB, j.SDB)
	if err != nil {
		log.Println("***>***> EventItemCreated: unexpected/unhandled error occurred when retriving entity on job. error:", err)
		return err
	}
	it, err := item.Retrieve(ctx, m.EntityID, m.ItemID, j.DB)
	if err != nil {
		log.Println("***>***> EventItemCreated: unexpected/unhandled error occurred while retriving item on job. error:", err)
		return err
	}

	valueAddedFields := e.ValueAdd(it.Fields())
	reference.UpdateChoicesWrapper(ctx, j.DB, j.SDB, m.AccountID, m.EntityID, valueAddedFields, NewJabEngine())

	ls, _ := stream.Retrieve(ctx, m.AccountID, m.ID, j.DB)
	m.State = ls.State

	//connect
	if m.State < stream.StateConnection {
		err = j.actOnConnections(m.AccountID, m.UserID, m.Source, m.EntityID, m.ItemID, valueAddedFields, nil, e.DisplayName, j.DB)
		if err != nil {
			log.Println("***>***> EventItemCreated: unexpected/unhandled error occurred on actOnConnections. error: ", err)
			return err
		} else {
			stream.Update(ctx, j.DB, m, "Connection", stream.StateConnection)
		}
	}

	//workflows
	if m.UserID != user.UUID_SYSTEM_USER && m.State < stream.StateWorkflow { // for now, preventing loops in workflows by this check!
		err = j.actOnWorkflows(ctx, e, m.ItemID, nil, it.Fields(), j.DB, j.SDB)
		if err != nil {
			log.Println("***>***> EventItemCreated: unexpected/unhandled error occurred on actOnWorkflows. error: ", err)
			return err
		} else {
			stream.Update(ctx, j.DB, m, "Workflow", stream.StateWorkflow)
		}
	}

	//categories such as email,meeting,members
	if m.State < stream.StateCategory {
		err = actOnCategories(ctx, m.AccountID, m.UserID, e, it, valueAddedFields, j.DB, j.SDB)
		if err != nil {
			log.Println("***>***> EventItemCreated: unexpected/unhandled error occurred on actOnIntegrations. error: ", err)
			return err
		} else {
			stream.Update(ctx, j.DB, m, "Category", stream.StateCategory)
		}
	}

	//who
	if m.State < stream.StateWho {
		err = j.actOnWho(m.AccountID, m.UserID, m.EntityID, m.ItemID, valueAddedFields, j.SDB)
		if err != nil {
			log.Println("***>***> EventItemCreated: unexpected/unhandled error occurred on actOnWho. error: ", err)
			return err
		} else {
			stream.Update(ctx, j.DB, m, "Who", stream.StateWho)
		}
	}

	//act on notifications
	if m.State < stream.StateNotification {
		err = j.actOnNotifications(ctx, m.AccountID, m.UserID, it.UpdatedAt, e, it.ID, it.UserID, nil, it.Fields(), m.Source, notification.TypeCreated)
		if err != nil {
			log.Println("***>***> EventItemCreated: unexpected/unhandled error occurred on notification update. error: ", err)
			return err
		} else {
			stream.Update(ctx, j.DB, m, "Notification", stream.StateNotification)
		}
	}

	//insertion in to redis graph DB
	//safely deleting the empty string...
	delete(m.Source, "")
	if m.State < stream.StateRedis {
		valueAddedFields = appendSystemProps(it, valueAddedFields)
		err = j.actOnRedisWrapper(ctx, m, valueAddedFields)
		if err != nil {
			return err
		}
	}
	//TODO delete the log stream.
	log.Println("***>***> Completed EventItemCreated ***<***<")
	return nil
}

func (j *Job) eventItemUpdated(m *stream.Message) error {
	log.Println("***>***> Reached EventItemUpdated ***<***<")
	ctx := context.Background()

	e, err := entity.Retrieve(ctx, m.AccountID, m.EntityID, j.DB, j.SDB)
	if err != nil {
		log.Println("***>***> EventItemUpdated: unexpected/unhandled error occurred when retriving entity inside job. error:", err)
		return err
	}

	it, err := item.Retrieve(ctx, m.EntityID, m.ItemID, j.DB)
	if err != nil {
		log.Println("***>***> EventItemUpdated: unexpected/unhandled error occurred while retriving item on job. error:", err)
		return err
	}

	valueAddedFields := e.ValueAdd(m.NewFields)

	ls, _ := stream.Retrieve(ctx, m.AccountID, m.ID, j.DB)
	m.State = ls.State

	//connections
	if m.State < stream.StateConnection {
		err = j.actOnConnections(m.AccountID, m.UserID, map[string][]string{}, m.EntityID, m.ItemID, valueAddedFields, e.ValueAdd(m.OldFields), e.DisplayName, j.DB)
		if err != nil {
			log.Println("***>***> EventItemUpdated: unexpected/unhandled error occurred on actOnConnections. error: ", err)
			return err
		} else {
			stream.Update(ctx, j.DB, m, "Connection", stream.StateConnection)
		}
	}

	//workflows
	if m.UserID != user.UUID_SYSTEM_USER && m.State < stream.StateWorkflow { // for now, preventing loops in workflows by this check!
		err = j.actOnWorkflows(ctx, e, m.ItemID, m.OldFields, m.NewFields, j.DB, j.SDB)
		if err != nil {
			log.Println("***>***> EventItemUpdated: unexpected/unhandled error occurred on actOnWorkflows. error: ", err)
			return err
		} else {
			stream.Update(ctx, j.DB, m, "Workflow", stream.StateWorkflow)
		}
	}

	//who
	if m.State < stream.StateWho {
		err = j.actOnWho(m.AccountID, m.UserID, m.EntityID, m.ItemID, valueAddedFields, j.SDB)
		if err != nil {
			log.Println("***>***> EventItemUpdated: unexpected/unhandled error occurred on actOnWho. error: ", err)
			return err
		} else {
			stream.Update(ctx, j.DB, m, "Who", stream.StateWho)
		}
	}

	//act on notifications
	if m.State < stream.StateNotification {
		err = j.actOnNotifications(ctx, m.AccountID, m.UserID, it.UpdatedAt, e, m.ItemID, it.UserID, m.OldFields, m.NewFields, m.Source, notification.TypeUpdated)
		if err != nil {
			log.Println("***>***> EventItemUpdated: unexpected/unhandled error occurred on notification update. error: ", err)
			return err
		} else {
			stream.Update(ctx, j.DB, m, "Notification", stream.StateNotification)
		}
	}

	//graph
	if m.State < stream.StateRedis {
		valueAddedFields = appendSystemProps(it, valueAddedFields)
		err = j.actOnRedisWrapper(ctx, m, valueAddedFields)
		if err != nil {
			return err
		}
	}
	//TODO delete the log stream.
	log.Println("***>***> Completed EventItemUpdated ***<***<")
	return nil
}

func (j *Job) eventItemReminded(m *stream.Message) error {
	ctx := context.Background()

	e, err := entity.Retrieve(ctx, m.AccountID, m.EntityID, j.DB, j.SDB)
	if err != nil {
		log.Println("***>***> EventItemReminded: unexpected/unhandled error occurred when retriving entity on job. error:", err)
		return err
	}
	it, err := item.Retrieve(ctx, m.EntityID, m.ItemID, j.DB)
	if err != nil {
		log.Println("***>***> EventItemReminded: unexpected/unhandled error occurred while retriving item on job. error:", err)
		return err
	}

	valueAddedFields := e.ValueAdd(it.Fields())
	reference.UpdateChoicesWrapper(ctx, j.DB, j.SDB, m.AccountID, m.EntityID, valueAddedFields, NewJabEngine())

	//act on notifications
	err = j.actOnNotifications(ctx, m.AccountID, m.UserID, it.UpdatedAt, e, it.ID, it.UserID, nil, it.Fields(), m.Source, notification.TypeReminder)
	if err != nil {
		log.Println("***>***> EventItemReminded: unexpected/unhandled error occurred on notification update. error: ", err)
		return err
	}
	return err
}

func (j *Job) eventItemDeleted(m *stream.Message) error {
	log.Println("***>***> Reached EventItemDeleted ***<***<")
	ctx := context.Background()

	e, err := entity.Retrieve(ctx, m.AccountID, m.EntityID, j.DB, j.SDB)
	if err != nil {
		log.Println("***>***> EventItemDeleted: unexpected/unhandled error occurred when retriving entity on job. error:", err)
		return err
	}

	it, err := item.Retrieve(ctx, m.EntityID, m.ItemID, j.DB)
	if err != nil {
		log.Println("***>***> EventItemDeleted: unexpected/unhandled error occurred while retriving item on job. error:", err)
		return err
	}

	ls, _ := stream.Retrieve(ctx, m.AccountID, m.ID, j.DB)
	m.State = ls.State

	err = destructOnIntegrations(ctx, m.AccountID, e, it, j.DB, j.SDB)
	if err != nil {
		log.Println("***>***> EventItemDeleted: unexpected/unhandled error occurred on destructOnIntegrations. error: ", err)
	}

	if m.State < stream.StatePrimaryDBDelete {
		err = item.Delete(ctx, j.DB, m.AccountID, m.EntityID, m.ItemID)
		if err != nil {
			log.Println("***>***> EventItemDeleted: unexpected/unhandled error occurred on delete main item. error: ", err)
			return err
		} else {
			stream.Update(ctx, j.DB, m, "Primary DB Delete", stream.StatePrimaryDBDelete)
		}
	}

	if m.State < stream.StateSecDBDelete {
		err = graphdb.Delete(j.SDB.GraphPool(), m.AccountID, m.EntityID, m.ItemID)
		if err != nil {
			log.Println("***>***> EventItemDeleted: unexpected/unhandled error occurred on delete redisgraph item. error: ", err)
			return err
		} else {
			stream.Update(ctx, j.DB, m, "Secondary DB Delete", stream.StateSecDBDelete)
		}
	}
	log.Println("***>***> Finished EventItemDeleted ***<***<")
	return nil
}

func (j *Job) eventEmailConvAdded(m *stream.Message) error {
	log.Println("***>***> Reached EventEmailConvAdded ***<***<")
	ctx := context.Background()

	cv, err := conversation.Retrieve(ctx, m.AccountID, m.ConversationID, j.DB)
	if err != nil {
		log.Println("***>***> EventEmailConvAdded: unexpected/unhandled error occurred while retriving item on job. error:", err)
		return err
	}

	emailEntity, err := entity.Retrieve(ctx, cv.AccountID, cv.EntityID, j.DB, j.SDB)
	if err != nil {
		log.Println("***>***> EventEmailConvAdded: unexpected/unhandled error occurred on retriving the entity on job. error:", err)
		return err
	}

	var emailItem entity.EmailEntity
	_, err = entity.RetrieveUnmarshalledItem(ctx, m.AccountID, m.EntityID, m.ItemID, &emailItem, j.DB, j.SDB)
	if err != nil {
		log.Println("***>***> EventEmailConvAdded: unexpected/unhandled error occurred on retriving the parent entity on job. error:", err)
		return err
	}

	//TODO push to job
	replyTo := emailItem.MessageID
	valueAddedFields := emailEntity.ValueAdd(cv.PayloadMap())
	msgID, err := email.SendMail(ctx, m.AccountID, m.EntityID, m.ItemID, valueAddedFields, replyTo, j.DB, j.SDB)
	if err != nil {
		log.Println("***>***> EventEmailConvAdded: unexpected/unhandled error occurred while sending mail. error:", err)
	}
	err = conversation.UpdateID(ctx, j.DB, cv.ID, *msgID, time.Now())
	if err != nil {
		log.Println("***>***> EventEmailConvAdded: unexpected/unhandled error occurred while updating the old conv id with msg id. error:", err)
	}
	log.Println("***>***> Finished EventEmailConvAdded ***<***<")
	return nil
}

func (j *Job) eventChatConvAdded(m *stream.Message) error {
	log.Println("***>***> Reached EventChatConvAdded ***<***<")
	ctx := context.Background()

	e, err := entity.Retrieve(ctx, m.AccountID, m.EntityID, j.DB, j.SDB)
	if err != nil {
		log.Println("***>***> EventChatConvAdded: unexpected/unhandled error occurred when retriving entity on job. error:", err)
		return err
	}
	it, err := item.Retrieve(ctx, m.EntityID, m.ItemID, j.DB)
	if err != nil {
		log.Println("***>***> EventChatConvAdded: unexpected/unhandled error occurred while retriving item on job. error:", err)
		return err
	}
	log.Println("***>***> Reaced FOR ITEM ***<***<", it.ID)

	cv, err := conversation.Retrieve(ctx, m.AccountID, m.ConversationID, j.DB)
	if err != nil {
		log.Println("***>***> EventEmailConvAdded: unexpected/unhandled error occurred while retriving item on job. error:", err)
		return err
	}

	valueAddedFields := e.ValueAdd(it.Fields())
	valueAddedFields = appendMessage(cv.Message, valueAddedFields) //passing message to act_on_notifications... valueadded --> item.Fields() --> dirtyFields() --> whoMessage

	ls, _ := stream.Retrieve(ctx, m.AccountID, m.ID, j.DB)
	m.State = ls.State

	//act on notifications
	if m.State < stream.StateNotification {
		err = j.actOnNotifications(ctx, m.AccountID, m.UserID, it.UpdatedAt, e, it.ID, &m.UserID, nil, entity.KeyValueMap(valueAddedFields), m.Source, notification.TypeChatConversationAdded)
		if err != nil {
			log.Println("***>***> EventChatConvAdded: unexpected/unhandled error occurred on notification update. error: ", err)
			return err
		} else {
			stream.Update(ctx, j.DB, m, "Notification", stream.StateNotification)
		}
	}

	//insertion in to redis graph DB
	//safely deleting the empty string...
	delete(m.Source, "")
	if m.State < stream.StateRedis {
		valueAddedFields = appendSystemProps(it, valueAddedFields)
		err = j.actOnRedisGraph(ctx, m.AccountID, m.EntityID, m.ItemID, nil, valueAddedFields, j.DB, j.SDB)
		if err != nil {
			log.Println("***>***> EventChatConvAdded: unexpected/unhandled error occurred on actOnRedisGraph. error: ", err)
			return err
		} else {
			stream.Update(ctx, j.DB, m, "Redis", stream.StateRedis)
		}
	}
	//TODO delete the log stream.
	log.Println("***>***> Completed EventChatConvAdded ***<***<")
	return nil
}

func (j *Job) eventEventAdded(m *stream.Message) error {
	log.Println("***>***> Reached EventEventAdded ***<***<")
	ctx := context.Background()
	if m.Source == nil {
		m.Source = make(map[string][]string, 0)
	}

	e, err := entity.Retrieve(ctx, m.AccountID, m.EntityID, j.DB, j.SDB)
	if err != nil {
		log.Println("***>***> EventEventCreated: unexpected/unhandled error occurred when retriving entity on job. error:", err)
		return err
	}
	ts, err := timeseries.Retrieve(ctx, m.AccountID, m.EntityID, m.ItemID, j.DB)
	if err != nil {
		log.Println("***>***> EventEventCreated: unexpected/unhandled error occurred while retriving item on job. error:", err)
		return err
	}

	if ts.Identifier != nil {
		identifierElements := strings.Split(*ts.Identifier, ":")
		if len(identifierElements) == 3 {
			conditionFields := make([]graphdb.Field, 0)
			entityName := identifierElements[0]
			fieldKey := identifierElements[1]
			fieldValue := identifierElements[1]

			e, err := entity.RetrieveByName(ctx, m.AccountID, entityName, j.DB)
			if err != nil {
				return err
			}
			exp := fmt.Sprintf("{{%s.%s}} eq {%s}", e.ID, fieldKey, fieldValue)
			filter := NewJabEngine().RunExpGrapher(ctx, j.DB, j.SDB, m.AccountID, exp)

			fields := e.OnlyVisibleFields()
			for _, f := range fields {
				if condition, ok := filter.Conditions[f.Key]; ok {
					conditionFields = append(conditionFields, f.MakeGraphField(condition.Term, condition.Expression, false))
				}
			}

			gSegment := graphdb.BuildGNode(m.AccountID, e.ID, false, nil).MakeBaseGNode("", conditionFields)
			result, err := graphdb.GetResult(j.SDB.GraphPool(), gSegment, 0, "", "")
			if err != nil {
				return err
			}

			m.Source[e.ID] = util.ParseGraphResultWithStrIDs(result)
		}
	}

	valueAddedFields := e.ValueAdd(ts.Fields())
	reference.UpdateChoicesWrapper(ctx, j.DB, j.SDB, m.AccountID, m.EntityID, valueAddedFields, NewJabEngine())

	ls, _ := stream.Retrieve(ctx, m.AccountID, m.ID, j.DB)
	m.State = ls.State

	//connect
	if m.State < stream.StateConnection {
		err = j.actOnConnections(m.AccountID, m.UserID, m.Source, m.EntityID, m.ItemID, valueAddedFields, e.ValueAdd(m.OldFields), e.DisplayName, j.DB)
		if err != nil {
			log.Println("***>***> EventItemCreated: unexpected/unhandled error occurred on actOnConnections. error: ", err)
			return err
		} else {
			stream.Update(ctx, j.DB, m, "Connection", stream.StateConnection)
		}
	}

	//insertion in to redis graph DB
	//safely deleting the empty string...
	delete(m.Source, "")
	if m.State < stream.StateRedis {
		err = j.actOnRedisWrapper(ctx, m, valueAddedFields)
		if err != nil {
			return err
		}
	}

	return nil
}

func (j *Job) EventUserSignedUp(accountName, emailAddress, draftID string, db *sqlx.DB, sdb *database.SecDB) error {
	ctx := context.Background()
	err := launchUser(ctx, draftID, accountName, "", "", emailAddress, db, sdb)
	if err != nil {
		log.Println("***>***> EventUserSignedUp: unexpected/unhandled error occurred while sending launch mail. error:", err)
		return err
	}
	return nil
}

func (j *Job) EventEmailReceived(db *sqlx.DB) {

}

func (j *Job) eventDelayExhausted(m *stream.Message) error {
	ctx := context.Background()
	triggerFlowID := m.Meta["trigger_flow_id"].(string)
	triggerNodeID := m.Meta["trigger_node_id"].(string)
	triggerEntityID := m.Meta["trigger_entity_id"].(string)
	triggerItemID := m.Meta["trigger_item_id"].(string)
	triggerFlowType := int(m.Meta["trigger_flow_type"].(float64))

	//removing it because its get added in the source inside the flow.
	delete(m.Meta, "trigger_flow_id")
	delete(m.Meta, "trigger_node_id")
	delete(m.Meta, "trigger_entity_id")
	delete(m.Meta, "trigger_item_id")
	delete(m.Meta, "trigger_flow_type")

	n, err := node.Retrieve(ctx, m.AccountID, triggerFlowID, triggerNodeID, j.DB)
	if err != nil {
		log.Println("***>***> EventDelayExhausted: unexpected error occurred on node retrive. error: ", err)
		return err
	} else {
		eng := engine.Engine{
			Job: j,
		}

		n.UpdateMeta(triggerEntityID, triggerItemID, triggerFlowType).UpdateVariables(triggerEntityID, triggerItemID)
		err = flow.StartJobFlow(ctx, j.DB, j.SDB, n, m.Meta, eng)
		if err != nil {
			log.Println("***>***> EventDelayExhausted: unexpected error occurred on startJobFlow. error: ", err)
			return err
		}
		return nil
	}
}

//act ons

func (j *Job) actOnRedisGraph(ctx context.Context, accountID, entityID, itemID string, oldFields map[string]interface{}, valueAddedFields []entity.Field, db *sqlx.DB, sdb *database.SecDB) error {
	log.Println("*********> debug internal.job actOnRedisGraph kicked in")
	if oldFields != nil { //use only during the update
		dirtyFields := item.Diff(oldFields, entity.KeyValueMap(valueAddedFields))

		//unlink
		for i := 0; i < len(valueAddedFields); i++ {
			f := &valueAddedFields[i]
			if newList, ok := dirtyFields[f.Key]; ok {
				if oldList, ok := oldFields[f.Key]; ok && f.Value != nil && (f.IsReference() || f.IsList()) {
					if oldList != nil && len(oldList.([]interface{})) > 0 {
						delList := deletedList(newList, oldList)
						f.UnlinkOffset = len(f.Value.([]interface{})) + 1
						f.Value = append(f.Value.(([]interface{})), delList...)
					}
				}
			}
		}
	}

	gpbNode := graphdb.BuildGNode(accountID, entityID, false, nil).MakeBaseGNode(itemID, makeGraphFields(valueAddedFields))
	if j.baseEntityID != "" && len(j.baseItemIDs) > 0 {

		relationShips, err := relationship.RetionshipType(ctx, db, accountID, j.baseEntityID, entityID)
		if err != nil {
			return errors.Wrap(err, "***> EventItemCreated: unexpected/unhandled error occurred when retriving relationships on job. error:")
		}

		connType := connectionType(j.baseEntityID, entityID, relationShips)

		for _, baseItemID := range j.baseItemIDs {
			switch connType {
			case 1: // one way reverse
				gpbNode = gpbNode.ParentEdge(j.baseEntityID, baseItemID, false) // contact creates companies : company has contacts
			case 2: // one way straight
				gpbNode = gpbNode.ParentEdge(j.baseEntityID, baseItemID, true) // contact creates companies : contact has companies
			case 3: // two way
				gpbNode = gpbNode.ParentEdge(j.baseEntityID, baseItemID, true)
				gpbNode = gpbNode.ParentEdge(j.baseEntityID, baseItemID, false)
			}
		}
	}

	err := graphdb.UpsertNode(sdb.GraphPool(), gpbNode)
	if err != nil {
		return errors.Wrap(err, "error: redisGrpah insertion job")
	}
	return nil
}

func (j *Job) actOnWorkflows(ctx context.Context, e entity.Entity, itemID string, oldFields, newFields map[string]interface{}, db *sqlx.DB, sdb *database.SecDB) error {
	log.Println("*********> debug internal.job actOnWorkflows kicked in")
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
				errs = flow.Trigger(ctx, db, sdb, itemID, dirtyFlows, eng)
			}
		case flow.FlowTypeEventCreate:
			errs = flow.Trigger(ctx, db, sdb, itemID, flows, eng)
		}
	}

	err = actOnPipelines(ctx, eng, e, itemID, dirtyFields, newFields, db, sdb)
	if err != nil && err != flow.ErrFlowActive {
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
func actOnPipelines(ctx context.Context, eng engine.Engine, e entity.Entity, itemID string, dirtyFields map[string]interface{}, newFields map[string]interface{}, db *sqlx.DB, sdb *database.SecDB) error {
	log.Println("*********> debug internal.job actOnPipelines kicked in")
	for _, fi := range e.EasyFields() {

		if dirtyField, ok := dirtyFields[fi.Key]; ok && fi.IsNode() && dirtyField != nil && len(dirtyField.([]interface{})) > 0 && fi.Dependent != nil {
			flowID := newFields[fi.Dependent.ParentKey].([]interface{})[0].(string)
			nodeID := dirtyField.([]interface{})[0].(string)
			err := flow.DirectTrigger(ctx, db, sdb, e.AccountID, flowID, nodeID, e.ID, itemID, eng)
			if err != nil {
				return errors.Wrap(err, "error: acting on pipelines")
			}
		}
	}
	return nil
}

// It connects the implicit relationships which as inferred by the field
// Right now, the connection helps fetching the events and nothing else. (check whether the flow/node gets added to redis because of this)
func (j Job) actOnConnections(accountID, userID string, base map[string][]string, entityID, itemID string, newFields, oldFields []entity.Field, entityName string, db *sqlx.DB) error {
	log.Println("*********> debug internal.job actOnConnections kicked in")
	ctx := context.Background()
	createEvent := oldFields == nil
	newValueAddedFieldsMap := entity.KeyMap(newFields)
	oldValueAddedFieldsMap := entity.KeyMap(oldFields)
	relationships, err := relationship.Relationships(ctx, db, accountID, entityID)
	if err != nil {
		return errors.Wrap(err, "error: querying relationships")
	}
	action := "updated"
	if createEvent {
		action = "created"
	}

	for _, r := range relationships {
		//Explicit connection happens when adding the item from inside the base element
		//The user can only delete that association and he couldn't update it becasue there is no
		//reference exists between the two entities implicitly.
		if r.FieldID == relationship.FieldAssociationKey && createEvent {
			if baseItemIDs, ok := base[r.DstEntityID]; ok { // same logic added to redis also
				for _, baseItemID := range baseItemIDs {
					err = connection.Associate(ctx, db, accountID, userID, r.RelationshipID, entityName, entityID, r.DstEntityID, itemID, baseItemID, newFields, action)
					if err != nil {
						return errors.Wrap(err, "error: querying association")
					}
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
						err := connection.Associate(ctx, db, accountID, userID, r.RelationshipID, entityName, entityID, f.RefID, itemID, dstItemID.(string), newFields, action)
						if err != nil {
							return errors.Wrap(err, "error: implicit connection with straight reference failed")
						}
					}
				}
			} else { //Implicit connection with reverse reference. When creating the contact inside a deal base
				//log.Println("internal.job implicit connection with reverse reference handled")
				if baseItemIDs, ok := base[r.DstEntityID]; ok && createEvent { //This won't happen during the update
					for _, baseItemID := range baseItemIDs {
						err = connection.Associate(ctx, db, accountID, userID, r.RelationshipID, entityName, entityID, r.DstEntityID, itemID, baseItemID, newFields, action)
						if err != nil {
							log.Println("***>***> actOnConnections: unexpected/unhandled error occurred when adding connections. error: ", err)
						}
						baseItem, err := item.Retrieve(ctx, r.DstEntityID, baseItemID, db)
						if err != nil {
							return errors.Wrap(err, "error: implicit connection with reverse reference failed")
						}
						itemFieldsMap := baseItem.Fields()
						log.Println("*********> debug internal.job BF itemFieldsMap ", itemFieldsMap)
						if vals, ok := itemFieldsMap[r.FieldID]; ok { // little complex
							exisitingVals := vals.([]interface{})
							exisitingVals = append(exisitingVals, itemID)
							itemFieldsMap[r.FieldID] = exisitingVals
							log.Println("*********> debug internal.job AF itemFieldsMap ", itemFieldsMap)
							_, err = item.UpdateFields(ctx, db, r.DstEntityID, baseItemID, itemFieldsMap)
							if err != nil {
								return errors.Wrap(err, "error: implicit connection with reverse reference failed")
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func actOnCategories(ctx context.Context, accountID, currentUserID string, e entity.Entity, it item.Item, valueAddedFields []entity.Field, db *sqlx.DB, sdb *database.SecDB) error {
	log.Println("*********> debug internal.job actOnCategories kicked in")
	//shall we move this to a common place
	acc, err := account.Retrieve(ctx, db, accountID)
	if err != nil {
		return err
	}

	userName := "System User"
	if currentUserID != user.UUID_SYSTEM_USER && currentUserID != user.UUID_ENGINE_USER && currentUserID != user.UUID_ANONYMOUS_USER {
		currentUser, err := user.RetrieveUser(ctx, db, currentUserID)
		if err != nil {
			return err
		}
		userName = *currentUser.Name
	}

	switch e.Category {
	case entity.CategoryEmail: //handles both receive and send
		var emailItem entity.EmailEntity
		entity.ParseFixedEntity(valueAddedFields, &emailItem)
		convType := conversation.TypeEmailReceived
		if emailItem.MessageID == "" { //sent
			convType = conversation.TypeEmailSent
			msgID, err := email.SendMail(ctx, accountID, e.ID, it.ID, valueAddedFields, "", db, sdb)
			if err != nil {
				return err
			}
			emailItem.MessageID = *msgID
		}
		//save the conversation
		err = conversation.SaveConversation(ctx, acc.ID, it.EntityID, it.ID, emailItem, emailItem.Body, convType, conversation.StateSent, db)
		if err != nil {
			return errors.Wrap(err, "unable to save conversation")
		}
	case entity.CategoryMeeting:
		err = calendar.CreateCalendarEvent(ctx, accountID, e.TeamID, e.ID, it.ID, valueAddedFields, db)
		if err == entity.ErrIntegNotFound {
			return nil
		}
	case entity.CategoryUsers:
		var usr entity.UserEntity
		jsonbody, _ := entity.MakeJSONBody(valueAddedFields)
		json.Unmarshal(jsonbody, &usr)
		teams, err := team.List(ctx, accountID, db)
		if err != nil {
			return err
		}
		err = notification.JoinInvitation(accountID, acc.Name, acc.Domain, team.Names(teams), userName, usr.Name, usr.Email, it.ID, db, sdb)
		if err != nil {
			return errors.Wrap(err, "unable to invite members")
		}
	}
	return err
}

func (j Job) actOnWho(accountID, userID, entityID, itemID string, valueAddedFields []entity.Field, sdb *database.SecDB) error {
	for _, f := range valueAddedFields {
		if f.Who == entity.WhoReminder && f.DataType == entity.TypeDateTime && f.Value != nil {
			when, err := util.ParseTime(f.Value.(string))
			if err != nil {
				return err
			}
			return (j).AddReminder(accountID, userID, entityID, itemID, when, sdb)
		}
	}
	return nil
}

func (j Job) actOnNotifications(ctx context.Context, accountID, userID string, itemUpdatedAt int64, e entity.Entity, itemID string, itemCreatorID *string, oldFields, newFields map[string]interface{}, source map[string][]string, notificationType notification.NotificationType) error {
	log.Println("*********> debug internal.job actOnNotifications kicked in")
	if e.Category == entity.CategoryNotification { // skip notification when a notification is created :P
		return nil
	}

	acc, err := account.Retrieve(ctx, j.DB, accountID)
	if err != nil {
		return err
	}

	dirtyFields := item.Diff(oldFields, newFields)
	//save the notification to the notifications.
	appNotifItem, err := notification.OnAnItemLevelEvent(ctx, userID, e.Category, e.DisplayName, acc.ID, acc.Domain, e.TeamID, e.ID, itemID, itemCreatorID, itemUpdatedAt, e.ValueAdd(newFields), dirtyFields, source, notificationType, j.DB, j.SDB, j.FirebaseSDKPath)
	if err != nil {
		return err
	}

	notifEntity, err := entity.Retrieve(ctx, appNotifItem.AccountID, appNotifItem.EntityID, j.DB, j.SDB)
	if err != nil {
		return err
	}
	valueAddedFields := notifEntity.ValueAdd(appNotifItem.Fields())
	valueAddedFields = appendSystemProps(*appNotifItem, valueAddedFields)
	err = j.actOnRedisGraph(ctx, appNotifItem.AccountID, appNotifItem.EntityID, appNotifItem.ID, nil, valueAddedFields, j.DB, j.SDB)
	if err != nil {
		return err
	}
	return nil
}

func destructOnIntegrations(ctx context.Context, accountID string, e entity.Entity, it item.Item, db *sqlx.DB, sdb *database.SecDB) error {
	var err error
	switch e.Category {
	case entity.CategoryEmailConfig:
		err = email.Destruct(ctx, accountID, e.ID, it.ID, db, sdb)
	case entity.CategoryCalendar:
		//calendar destruct yet to be implemented
	}
	return err
}

func connectionType(baseEntityID, entityID string, relationShips []relationship.Relationship) int {
	typeOfConnection := 0 // no connection

	for _, r := range relationShips {
		if r.SrcEntityID == entityID && r.DstEntityID == baseEntityID {
			if typeOfConnection == 2 {
				typeOfConnection = 3 // two way connection
			} else {
				typeOfConnection = 1 // one way reverse
			}

		} else if r.DstEntityID == entityID && r.SrcEntityID == baseEntityID {
			if typeOfConnection == 1 {
				typeOfConnection = 3 // two way connection
			} else {
				typeOfConnection = 2 // one way staright
			}

		}
	}
	return typeOfConnection
}

func appendSystemProps(it item.Item, valueAddedFields []entity.Field) []entity.Field {

	createdAtField := entity.Field{
		Key:   "system_created_at",
		Value: util.GetMilliSecondsFloat(it.CreatedAt),
	}
	updatedAtField := entity.Field{
		Key:   "system_updated_at",
		Value: util.GetMilliSecondsFloat(util.ConvertMilliToTime(it.UpdatedAt)),
	}
	isPubilcField := entity.Field{
		Key:   "system_is_public",
		Value: it.IsPublic,
	}
	valueAddedFields = append(valueAddedFields, createdAtField, updatedAtField, isPubilcField)
	if it.UserID != nil {
		createdByField := entity.Field{
			Key:   "system_created_by",
			Value: *it.UserID,
		}
		valueAddedFields = append(valueAddedFields, createdByField)
	}

	return valueAddedFields
}

func appendMessage(message string, valueAddedFields []entity.Field) []entity.Field {
	messageField := entity.Field{
		Key:   entity.WhoMessage,
		Value: util.TruncateText(message, 20),
	}

	valueAddedFields = append(valueAddedFields, messageField)
	return valueAddedFields
}

func deletedList(newList, oldList interface{}) []interface{} {
	newMap := make(map[string]bool, 0)
	for _, n := range newList.([]interface{}) {
		newMap[n.(string)] = true
	}

	deletedList := make([]interface{}, 0)
	//deletedList = append(deletedList, "--") //TODO: hacky-none-fix
	for i := 0; i < len(oldList.([]interface{})); i++ {
		o := oldList.([]interface{})[i]
		if _, ok := newMap[o.(string)]; !ok {
			deletedList = append(deletedList, o)
		}
	}
	return deletedList
}

func (j *Job) actOnRedisWrapper(ctx context.Context, m *stream.Message, valueAddedFields []entity.Field) error {
	if len(m.Source) == 0 {
		err := j.actOnRedisGraph(ctx, m.AccountID, m.EntityID, m.ItemID, m.OldFields, valueAddedFields, j.DB, j.SDB)
		if err != nil {
			return err
		} else {
			stream.Update(ctx, j.DB, m, "Redis", stream.StateRedis)
		}
	} else {
		if m.OldFields != nil { //for update case need not loop base
			err := j.actOnRedisGraph(ctx, m.AccountID, m.EntityID, m.ItemID, m.OldFields, valueAddedFields, j.DB, j.SDB)
			if err != nil {
				return err
			}
		} else {
			for j.baseEntityID, j.baseItemIDs = range m.Source {
				if j.baseEntityID == "" {
					continue
				}
				err := j.actOnRedisGraph(ctx, m.AccountID, m.EntityID, m.ItemID, m.OldFields, valueAddedFields, j.DB, j.SDB)
				if err != nil {
					return err
				}
			}
		}
		stream.Update(ctx, j.DB, m, "Redis", stream.StateRedis)
	}
	return nil
}

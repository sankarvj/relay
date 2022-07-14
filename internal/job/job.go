package job

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/account"
	"gitlab.com/vjsideprojects/relay/internal/connection"
	conv "gitlab.com/vjsideprojects/relay/internal/conversation"
	"gitlab.com/vjsideprojects/relay/internal/discovery"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/integration/calendar"
	"gitlab.com/vjsideprojects/relay/internal/integration/email"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/notification"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/stream"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/reference"
	"gitlab.com/vjsideprojects/relay/internal/relationship"
	"gitlab.com/vjsideprojects/relay/internal/rule/engine"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"gitlab.com/vjsideprojects/relay/internal/user"
)

//func's in this package should not throw errors. It should handle errors by re-queue/dl-queue
type Job struct {
	baseItemID      string
	baseEntityID    string
	DB              *sqlx.DB
	Rpool           *redis.Pool
	FirebaseSDKPath string
}

func NewJob(db *sqlx.DB, rp *redis.Pool, firebaseSDKPath string) *Job {
	return &Job{DB: db, Rpool: rp, FirebaseSDKPath: firebaseSDKPath}
}

func (j *Job) Post(msg *stream.Message) error {
	switch msg.Type {
	case stream.TypeItemCreate:
		j.eventItemCreated(*msg)
	case stream.TypeItemUpdate:
		j.eventItemUpdated(*msg)
	case stream.TypeItemDelete:
		j.eventItemDeleted(*msg)
	case stream.TypeItemRemind:
		j.eventItemReminded(*msg)
	case stream.TypeItemDelayed:
		j.eventDelayExhausted(*msg)
	case stream.TypeConversationAdded:
		j.eventConvAdded(*msg)
	}
	return nil
}

// events

func (j *Job) eventItemUpdated(m stream.Message) {
	log.Println("***>***> Reached EventItemUpdated ***<***<")
	ctx := context.Background()

	e, err := entity.Retrieve(ctx, m.AccountID, m.EntityID, j.DB)
	if err != nil {
		log.Println("***>***> EventItemUpdated: unexpected/unhandled error occurred when retriving entity inside job. error:", err)
		return
	}

	it, err := item.Retrieve(ctx, m.EntityID, m.ItemID, j.DB)
	if err != nil {
		log.Println("***>***> EventItemUpdated: unexpected/unhandled error occurred while retriving item on job. error:", err)
		return
	}

	valueAddedFields := e.ValueAdd(m.NewFields)

	//workflows
	if m.UserID != engine.UUID_SYSTEM_USER { // for now, preventing loops in workflows by this check!
		err = j.actOnWorkflows(ctx, e, m.ItemID, m.OldFields, m.NewFields, j.DB, j.Rpool)
		if err != nil {
			log.Println("***>***> EventItemUpdated: unexpected/unhandled error occurred on actOnWorkflows. error: ", err)
			return
		}
	}

	//connections
	err = j.actOnConnections(m.AccountID, m.UserID, map[string]string{}, m.EntityID, m.ItemID, valueAddedFields, e.ValueAdd(m.OldFields), e.DisplayName, "updated", j.DB)
	if err != nil {
		log.Println("***>***> EventItemUpdated: unexpected/unhandled error occurred on actOnConnections. error: ", err)
		return
	}

	//who
	err = j.actOnWho(m.AccountID, m.UserID, m.EntityID, m.ItemID, valueAddedFields, j.Rpool)
	if err != nil {
		log.Println("***>***> EventItemUpdated: unexpected/unhandled error occurred on actOnWho. error: ", err)
		return
	}

	//act on notifications
	err = j.actOnNotifications(ctx, m.AccountID, m.UserID, e, m.ItemID, it.UserID, m.OldFields, m.NewFields, m.Source, notification.TypeUpdated)
	if err != nil {
		log.Println("***>***> EventItemUpdated: unexpected/unhandled error occurred on notification update. error: ", err)
	}

	//graph
	valueAddedFields = appendTimers(it.CreatedAt, util.ConvertMilliToTime(it.UpdatedAt), *it.UserID, valueAddedFields)
	err = j.actOnRedisGraph(ctx, m.AccountID, m.EntityID, m.ItemID, m.OldFields, valueAddedFields, j.DB, j.Rpool)
	if err != nil {
		log.Println("***>***> EventItemUpdated: unexpected/unhandled error occurred on actOnRedisGraph. error: ", err)
		return
	}
}

func (j *Job) eventItemCreated(m stream.Message) {
	log.Println("***>***> Reached EventItemCreated ***<***<")
	ctx := context.Background()

	e, err := entity.Retrieve(ctx, m.AccountID, m.EntityID, j.DB)
	if err != nil {
		log.Println("***>***> EventItemCreated: unexpected/unhandled error occurred when retriving entity on job. error:", err)
		return
	}
	it, err := item.Retrieve(ctx, m.EntityID, m.ItemID, j.DB)
	if err != nil {
		log.Println("***>***> EventItemCreated: unexpected/unhandled error occurred while retriving item on job. error:", err)
		return
	}

	valueAddedFields := e.ValueAdd(it.Fields())
	reference.UpdateChoicesWrapper(ctx, j.DB, m.AccountID, m.EntityID, valueAddedFields, NewJabEngine())

	//workflows
	if m.UserID != engine.UUID_SYSTEM_USER { // for now, preventing loops in workflows by this check!
		err = j.actOnWorkflows(ctx, e, m.ItemID, nil, it.Fields(), j.DB, j.Rpool)
		if err != nil {
			log.Println("***>***> EventItemCreated: unexpected/unhandled error occurred on actOnWorkflows. error: ", err)
			return
		}
	}

	//connect
	err = j.actOnConnections(m.AccountID, m.UserID, m.Source, m.EntityID, m.ItemID, valueAddedFields, nil, e.DisplayName, "created", j.DB)
	if err != nil {
		log.Println("***>***> EventItemCreated: unexpected/unhandled error occurred on actOnConnections. error: ", err)
		return
	}

	//categories such as email,meeting,members
	err = actOnCategories(ctx, m.AccountID, m.UserID, e, it, valueAddedFields, j.DB, j.Rpool)
	if err != nil {
		log.Println("***>***> EventItemCreated: unexpected/unhandled error occurred on actOnIntegrations. error: ", err)
		return
	}

	//who
	err = j.actOnWho(m.AccountID, m.UserID, m.EntityID, m.ItemID, valueAddedFields, j.Rpool)
	if err != nil {
		log.Println("***>***> EventItemCreated: unexpected/unhandled error occurred on actOnWho. error: ", err)
		return
	}

	//act on notifications
	err = j.actOnNotifications(ctx, m.AccountID, m.UserID, e, it.ID, it.UserID, nil, it.Fields(), m.Source, notification.TypeCreated)
	if err != nil {
		log.Println("***>***> EventItemCreated: unexpected/unhandled error occurred on notification update. error: ", err)
	}

	valueAddedFields = appendTimers(it.CreatedAt, util.ConvertMilliToTime(it.UpdatedAt), *it.UserID, valueAddedFields)
	//insertion in to redis graph DB
	//safely deleting the empty string...
	delete(m.Source, "")
	if len(m.Source) == 0 {
		err = j.actOnRedisGraph(ctx, m.AccountID, m.EntityID, m.ItemID, nil, valueAddedFields, j.DB, j.Rpool)
		if err != nil {
			log.Println("***>***> EventItemCreated: unexpected/unhandled error occurred on actOnRedisGraph. error: ", err)
			return
		}
	} else {
		log.Printf("m.Source %+v", m.Source)
		for j.baseEntityID, j.baseItemID = range m.Source {

			if j.baseEntityID == "" {
				log.Println("***>***> EventItemCreated: unexpected/unhandled error occurred on actOnRedisGraph: baseEntityID is empty")
				continue
			}
			err = j.actOnRedisGraph(ctx, m.AccountID, m.EntityID, m.ItemID, nil, valueAddedFields, j.DB, j.Rpool)
			if err != nil {
				log.Println("***>***> EventItemCreated: unexpected/unhandled error occurred on actOnRedisGraph. error: ", err)
				return
			}
		}
	}
	log.Println("***>***> Completed EventItemCreated ***<***<")
}

func (j *Job) eventItemReminded(m stream.Message) {
	ctx := context.Background()

	e, err := entity.Retrieve(ctx, m.AccountID, m.EntityID, j.DB)
	if err != nil {
		log.Println("***>***> EventItemReminded: unexpected/unhandled error occurred when retriving entity on job. error:", err)
		return
	}
	it, err := item.Retrieve(ctx, m.EntityID, m.ItemID, j.DB)
	if err != nil {
		log.Println("***>***> EventItemReminded: unexpected/unhandled error occurred while retriving item on job. error:", err)
		return
	}

	valueAddedFields := e.ValueAdd(it.Fields())
	reference.UpdateChoicesWrapper(ctx, j.DB, m.AccountID, m.EntityID, valueAddedFields, NewJabEngine())

	//act on notifications
	err = j.actOnNotifications(ctx, m.AccountID, m.UserID, e, it.ID, it.UserID, nil, it.Fields(), m.Source, notification.TypeReminder)
	if err != nil {
		log.Println("***>***> EventItemReminded: unexpected/unhandled error occurred on notification update. error: ", err)
	}
}

func (j *Job) eventItemDeleted(m stream.Message) {
	log.Println("***>***> Reached EventItemDeleted ***<***<")
	ctx := context.Background()

	e, err := entity.Retrieve(ctx, m.AccountID, m.EntityID, j.DB)
	if err != nil {
		log.Println("***>***> EventItemDeleted: unexpected/unhandled error occurred when retriving entity on job. error:", err)
		return
	}

	it, err := item.Retrieve(ctx, m.EntityID, m.ItemID, j.DB)
	if err != nil {
		log.Println("***>***> EventItemDeleted: unexpected/unhandled error occurred while retriving item on job. error:", err)
		return
	}

	err = destructOnIntegrations(ctx, m.AccountID, e, it, j.DB)
	if err != nil {
		log.Println("***>***> EventItemDeleted: unexpected/unhandled error occurred on destructOnIntegrations. error: ", err)
		return
	}

	err = item.Delete(ctx, j.DB, m.AccountID, m.EntityID, m.ItemID)
	if err != nil {
		log.Println("***>***> EventItemDeleted: unexpected/unhandled error occurred on delete main item. error: ", err)
		return
	}

	err = graphdb.Delete(j.Rpool, m.AccountID, m.EntityID, m.ItemID)
	if err != nil {
		log.Println("***>***> EventItemDeleted: unexpected/unhandled error occurred on delete redisgraph item. error: ", err)
		return
	}
	log.Println("***>***> Finished EventItemDeleted ***<***<")
}

func (j *Job) eventConvAdded(m stream.Message) {
	ctx := context.Background()
	e, err := entity.Retrieve(ctx, m.AccountID, m.EntityID, j.DB)
	if err != nil {
		log.Println("***>***> EventConvAdded: unexpected/unhandled error occurred on retriving the entity on job. error:", err)
		return
	}

	var parentEmailEntityItem entity.EmailEntity
	_, err = entity.RetrieveUnmarshalledItem(ctx, m.AccountID, m.EntityID, m.ItemID, &parentEmailEntityItem, j.DB)
	if err != nil {
		log.Println("***>***> EventConvAdded: unexpected/unhandled error occurred on retriving the parent entity on job. error:", err)
		return
	}

	cv, err := conv.Retrieve(ctx, m.AccountID, m.ConversationID, j.DB)
	if err != nil {
		log.Println("***>***> EventConvAdded: unexpected/unhandled error occurred while retriving item on job. error:", err)
		return
	}

	//TODO push to job
	replyTo := parentEmailEntityItem.MessageID
	valueAddedFields := e.ValueAdd(cv.PayloadMap())
	_, err = email.SendMail(ctx, m.AccountID, m.EntityID, m.ItemID, valueAddedFields, replyTo, j.DB)
	if err != nil {
		log.Println("***>***> EventConvAdded: unexpected/unhandled error occurred while sending mail. error:", err)
	}
}

func (j *Job) EventUserSignedUp(accountName, emailAddress, draftID string, db *sqlx.DB, rp *redis.Pool) {
	err := launchUser(draftID, accountName, "", "", emailAddress, db, rp)
	if err != nil {
		log.Println("***>***> EventUserSignedUp: unexpected/unhandled error occurred while sending launch mail. error:", err)
	}
}

func (j *Job) EventEmailReceived(db *sqlx.DB) {

}

func (j *Job) eventDelayExhausted(m stream.Message) {
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
	} else {
		eng := engine.Engine{
			Job: j,
		}

		n.UpdateMeta(triggerEntityID, triggerItemID, triggerFlowType).UpdateVariables(triggerEntityID, triggerItemID)
		err = flow.StartJobFlow(ctx, j.DB, j.Rpool, n, m.Meta, eng)
		if err != nil {
			log.Println("***>***> EventDelayExhausted: unexpected error occurred on startJobFlow. error: ", err)
		}
	}
}

//act ons

func (j *Job) actOnRedisGraph(ctx context.Context, accountID, entityID, itemID string, oldFields map[string]interface{}, valueAddedFields []entity.Field, db *sqlx.DB, rp *redis.Pool) error {
	log.Println("*********> debug internal.job actOnRedisGraph kicked in")
	if oldFields != nil { //use only during the update
		dirtyFields := item.Diff(oldFields, entity.FieldsMap(valueAddedFields))

		//unlink
		for i := 0; i < len(valueAddedFields); i++ {
			f := &valueAddedFields[i]
			if _, ok := dirtyFields[f.Key]; ok {
				if old, ok := oldFields[f.Key]; ok && f.Value != nil && (f.IsReference() || f.IsList()) {
					if len(old.([]interface{})) > 0 {
						f.UnlinkOffset = len(f.Value.([]interface{})) + 1
						f.Value = append(f.Value.(([]interface{})), old.([]interface{})...)
					}
				}
			}
		}
	}

	gpbNode := graphdb.BuildGNode(accountID, entityID, false).MakeBaseGNode(itemID, makeGraphFields(valueAddedFields))
	if j.baseEntityID != "" && j.baseItemID != "" {

		relationShips, err := relationship.RetionshipType(ctx, db, accountID, j.baseEntityID, entityID)
		if err != nil {
			return errors.Wrap(err, "***> EventItemCreated: unexpected/unhandled error occurred when retriving relationships on job. error:")
		}

		connType := connectionType(j.baseEntityID, entityID, relationShips)
		switch connType {
		case 1: // one way reverse
			gpbNode = gpbNode.ParentEdge(j.baseEntityID, j.baseItemID, false) // contact creates companies : company has contacts
		case 2: // one way straight
			gpbNode = gpbNode.ParentEdge(j.baseEntityID, j.baseItemID, true) // contact creates companies : contact has companies
		case 3: // two way
			gpbNode = gpbNode.ParentEdge(j.baseEntityID, j.baseItemID, true)
			gpbNode = gpbNode.ParentEdge(j.baseEntityID, j.baseItemID, false)
		}

	}

	err := graphdb.UpsertNode(rp, gpbNode)
	if err != nil {
		return errors.Wrap(err, "error: redisGrpah insertion job")
	}
	return nil
}

func (j *Job) actOnWorkflows(ctx context.Context, e entity.Entity, itemID string, oldFields, newFields map[string]interface{}, db *sqlx.DB, rp *redis.Pool) error {
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
	log.Println("*********> debug internal.job actOnPipelines kicked in")
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
// Right now, the connection helps fetching the events and nothing else. (check whether the flow/node gets added to redis because of this)
func (j Job) actOnConnections(accountID, userID string, base map[string]string, entityID, itemID string, newFields, oldFields []entity.Field, entityName, action string, db *sqlx.DB) error {
	log.Println("*********> debug internal.job actOnConnections kicked in")
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
			if baseItemID, ok := base[r.DstEntityID]; ok { // same logic added to redis also
				err = connection.Associate(ctx, db, accountID, userID, r.RelationshipID, entityName, entityID, r.DstEntityID, itemID, baseItemID, newFields, action)
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
						err := connection.Associate(ctx, db, accountID, userID, r.RelationshipID, entityName, entityID, f.RefID, itemID, dstItemID.(string), newFields, action)
						if err != nil {
							return errors.Wrap(err, "error: implicit connection with straight reference failed")
						}
					}
				}
			} else { //Implicit connection with reverse reference. When creating the contact inside a deal base
				//log.Println("internal.job implicit connection with reverse reference handled")
				if baseItemID, ok := base[r.DstEntityID]; ok && createEvent { //This won't happen during the update
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
	return nil
}

func actOnCategories(ctx context.Context, accountID, currentUserID string, e entity.Entity, it item.Item, valueAddedFields []entity.Field, db *sqlx.DB, rp *redis.Pool) error {
	log.Println("*********> debug internal.job actOnCategories kicked in")
	//shall we move this to a common place
	acc, err := account.Retrieve(ctx, db, accountID)
	if err != nil {
		return err
	}

	userName := "System User"
	if currentUserID != engine.UUID_SYSTEM_USER {
		currentUser, err := user.RetrieveUser(ctx, db, currentUserID)
		if err != nil {
			return err
		}
		userName = *currentUser.Name
	}

	switch e.Category {
	case entity.CategoryEmail:
		if it.Name == nil || *it.Name != "received" { //super hacky :( Trying to avoid the sendmail action when saving the received mail
			msgID, err := email.SendMail(ctx, accountID, e.ID, it.ID, valueAddedFields, "", db)
			if err == nil {
				err = saveMsgID(ctx, accountID, e.ID, it.ID, *msgID, db)
				if err != nil {
					log.Println("***>***> actOnConnections: unexpected/unhandled error occurred when sending mails. error: ", err)
				}
			}
		}
	case entity.CategoryMeeting:
		err = calendar.CreateCalendarEvent(ctx, accountID, e.TeamID, e.ID, it.ID, valueAddedFields, db)
	case entity.CategoryUsers:
		var usr entity.UserEntity
		jsonbody, _ := entity.MakeJSONBody(valueAddedFields)
		json.Unmarshal(jsonbody, &usr)
		err = notification.JoinInvitation(accountID, acc.Name, userName, usr.Name, usr.Email, it.ID, db, rp)
	}
	return err
}

func (j Job) actOnWho(accountID, userID, entityID, itemID string, valueAddedFields []entity.Field, rp *redis.Pool) error {
	for _, f := range valueAddedFields {
		if f.Who == entity.WhoReminder && f.DataType == entity.TypeDateTime && f.Value != nil {
			when, err := util.ParseTime(f.Value.(string))
			if err != nil {
				return err
			}
			return (j).AddReminder(accountID, userID, entityID, itemID, when, rp)
		}
	}
	return nil
}

func (j Job) actOnNotifications(ctx context.Context, accountID, userID string, e entity.Entity, itemID string, itemCreatorID *string, oldFields, newFields map[string]interface{}, source map[string]string, notificationType notification.NotificationType) error {
	log.Println("*********> debug internal.job actOnNotifications kicked in")
	baseIds := make([]string, 0)
	for baseEntityID, baseItemID := range source {
		baseID := fmt.Sprintf("%s#%s", baseEntityID, baseItemID)
		baseIds = append(baseIds, baseID)
	}

	if e.Category == entity.CategoryNotification {
		return nil
	}
	dirtyFields := item.Diff(oldFields, newFields)
	//save the notification to the notifications.
	notifItem, err := notification.OnAnItemLevelEvent(ctx, userID, e.DisplayName, accountID, e.TeamID, e.ID, itemID, itemCreatorID, e.ValueAdd(newFields), dirtyFields, baseIds, notificationType, j.DB, j.FirebaseSDKPath)
	if err != nil {
		return err
	}

	notifEntity, err := entity.Retrieve(ctx, notifItem.AccountID, notifItem.EntityID, j.DB)
	if err != nil {
		return err
	}
	valueAddedFields := notifEntity.ValueAdd(notifItem.Fields())

	valueAddedFields = appendTimers(notifItem.CreatedAt, util.ConvertMilliToTime(notifItem.UpdatedAt), *notifItem.UserID, valueAddedFields)
	err = j.actOnRedisGraph(ctx, notifItem.AccountID, notifItem.EntityID, notifItem.ID, nil, valueAddedFields, j.DB, j.Rpool)
	if err != nil {
		return err
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
	log.Println("typeOfConnection ", typeOfConnection)
	return typeOfConnection
}

func appendTimers(createdAt, updatedAt time.Time, userID string, valueAddedFields []entity.Field) []entity.Field {
	createdAtField := entity.Field{
		Key:   "system_created_at",
		Value: util.GetMilliSecondsFloat(createdAt),
	}
	updatedAtField := entity.Field{
		Key:   "system_updated_at",
		Value: util.GetMilliSecondsFloat(updatedAt),
	}
	createdByField := entity.Field{
		Key:   "system_created_by",
		Value: userID,
	}
	valueAddedFields = append(valueAddedFields, createdAtField, updatedAtField, createdByField)
	return valueAddedFields
}

package job

import (
	"context"
	"encoding/base64"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/connection"
	"gitlab.com/vjsideprojects/relay/internal/draft"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/notification"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/ruler"
	"go.opencensus.io/trace"
	"golang.org/x/crypto/bcrypt"
)

//TODO can be removed. not used anywhere
func createActivityEvent(ctx context.Context, baseItemID string, ae entity.Entity, childEntity entity.Entity, childItem item.Item, db *sqlx.DB) (item.Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.event.Create")
	defer span.End()

	ni := item.NewItem{}
	ni.ID = uuid.New().String()
	ni.AccountID = ae.AccountID
	ni.EntityID = ae.ID
	ni.UserID = childItem.UserID
	ni.GenieID = &baseItemID
	ni.Fields = make(map[string]interface{}, 0)

	actualItemFields := childEntity.ValueAdd(childItem.Fields())
	namedActualFields := entity.MetaMap(actualItemFields)

	activityFields := ae.EasyFields()
	namedActivityFields := entity.NameMap(activityFields)

	ni.Fields[namedActivityFields["activity-name"].Key] = childEntity.Name
	ni.Fields[namedActivityFields["activity-action"].Key] = namedActualFields["title"].Value
	ni.Fields[namedActivityFields["activity-link"].Key] = ""

	evItem, err := item.Create(ctx, db, ni, time.Now())
	if err != nil {
		return evItem, err
	}

	return evItem, nil
}

func launchUser(ctx context.Context, draftID, accountName, requester, usrName, usrEmail string, db *sqlx.DB, sdb *database.SecDB) error {
	dr, err := draft.Retrieve(ctx, draftID, db)
	if err != nil {
		return err
	}
	return notification.WelcomeInvitation(draftID, dr.Teams, accountName, dr.Host, requester, usrName, usrEmail, db, sdb)
}

func compare(ctx context.Context, db *sqlx.DB, accountID, relationshipID string, f, of entity.Field) []interface{} {
	if ruler.Compare(f.Value, of.Value) { // handle delete alone here
		deletedItems, newItems := item.CompareItems(f.Value.([]interface{}), of.Value.([]interface{}))
		for _, deletedItem := range deletedItems {
			err := connection.Delete(ctx, db, relationshipID, deletedItem.(string))
			if err != nil {
				log.Println("***> unexpected/unhandled error occurred when deleting connection. error:", err)
			}
		}
		return newItems
	}
	return []interface{}{}
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
		Key:          f.Key,
		Value:        f.Value,
		DataType:     graphdb.DType(f.DataType),
		RefID:        f.RefID,
		Field:        makeGraphField(f.Field),
		UnlinkOffset: f.UnlinkOffset,
	}
}

func emailHash(emailAddress string) (string, error) {
	bmHash, err := bcrypt.GenerateFromPassword([]byte(emailAddress), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bmHash), nil
}

func (j Job) kabali(ctx context.Context, accountID, teamID, userID, entityID, itemID string, approverField entity.Field, valueAddedFields []entity.Field, db *sqlx.DB, sdb *database.SecDB) error {
	log.Println("TODO: Bad Logic To Handle Adding Approvals To Task. These logic definetly needs to be revisited")
	var dueByVal interface{}
	for _, vaf := range valueAddedFields {
		if vaf.Who == entity.WhoDueBy {
			dueByVal = vaf.Value
		}
	}

	// add a approver under current item
	approvalEntity, err := entity.RetrieveFixedEntity(ctx, db, accountID, teamID, entity.FixedEntityApprovals)
	if err != nil {
		return err
	}
	itemFieldsMap := make(map[string]interface{}, 0)
	approvalFields := approvalEntity.EasyFields()
	for _, apf := range approvalFields {
		if apf.Who == entity.WhoAssignee {
			itemFieldsMap[apf.Key] = approverField.Value
		} else if apf.Who == entity.WhoStatus {
			refItems, _ := item.EntityItems(ctx, accountID, apf.RefID, db)
			if len(refItems) > 0 {
				itemFieldsMap[apf.Key] = []interface{}{refItems[0].ID}
			}
		} else if apf.Who == entity.WhoDueBy {
			if dueByVal != nil {
				itemFieldsMap[apf.Key] = dueByVal
			}
		}
	}
	ni := item.NewItem{
		ID:        uuid.New().String(),
		AccountID: accountID,
		GenieID:   &itemID,
		UserID:    &userID,
		EntityID:  approvalEntity.ID,
		Fields:    itemFieldsMap,
		Source:    map[string][]string{entityID: {itemID}},
	}

	it, err := item.Create(ctx, db, ni, time.Now())
	if err != nil {
		return err
	}

	existingBE := j.baseEntityID
	existingBIDS := j.baseItemIDs

	j.baseEntityID = entityID
	j.baseItemIDs = []string{itemID}
	err = j.actOnRedisGraph(ctx, accountID, it.EntityID, it.ID, nil, approvalEntity.ValueAdd(it.Fields()), db, sdb)
	if err != nil {
		j.baseEntityID = existingBE
		j.baseItemIDs = existingBIDS
		return err
	}
	j.baseEntityID = existingBE
	j.baseItemIDs = existingBIDS

	return nil
}

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

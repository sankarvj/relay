package job

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/connection"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/notification"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	eml "gitlab.com/vjsideprojects/relay/internal/platform/integration/email"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/ruler"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
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
	namedActualFields := entity.MetaFieldsObjMap(actualItemFields)

	activityFields := ae.FieldsIgnoreError()
	namedActivityFields := entity.NamedFieldsObjMap(activityFields)

	ni.Fields[namedActivityFields["activity-name"].Key] = childEntity.Name
	ni.Fields[namedActivityFields["activity-action"].Key] = namedActualFields["title"].Value
	ni.Fields[namedActivityFields["activity-link"].Key] = ""

	evItem, err := item.Create(ctx, db, ni, time.Now())
	if err != nil {
		return evItem, err
	}

	return evItem, nil
}

func inviteUser(accountID, accountName, requester, usrName, usrEmail string, db *sqlx.DB, rp *redis.Pool) error {
	ctx := context.Background()
	err := notification.UserInvitation(ctx)
	if err != nil {
		return err
	}

	magicLink, err := auth.CreateMagicLink(accountID, usrEmail, rp)
	if err != nil {
		log.Println("***>***> EventUserInvited: unexpected/unhandled error occurred when creating the magic link. error:", err)
		return err
	}

	toField := []interface{}{usrEmail}
	subject := fmt.Sprintf("%s has invited you to join %s", requester, accountName)
	body := fmt.Sprintf("Hi %s, You are invited to join %s. Please click this <a href='%s'>link</a> to get started", usrName, accountName, magicLink)
	e := eml.FallbackMail{Domain: "", ReplyTo: ""}
	_, err = e.SendMail("", "contact@wayplot.com", "", util.ConvertSliceTypeRev(toField), subject, body)
	if err != nil {
		log.Println("***>***> EventUserInvited: unexpected/unhandled error occurred when user invited to join account. error:", err)
		return err
	}
	return nil
}

func launchUser(draftID, accountName, requester, usrName, usrEmail string, db *sqlx.DB, rp *redis.Pool) error {
	ctx := context.Background()
	err := notification.UserInvitation(ctx)
	if err != nil {
		return err
	}

	magicLink, err := auth.CreateMagicLaunchLink(draftID, accountName, usrEmail, rp)
	if err != nil {
		log.Println("***>***> EventUserInvited: unexpected/unhandled error occurred when creating the magic link. error:", err)
		return err
	}

	toField := []interface{}{usrEmail}
	subject := fmt.Sprintf("%s is ready", accountName)
	body := fmt.Sprintf("Hi, /n Your account is ready. Please click this <a href='%s'>magic link</a> to start", magicLink)
	e := eml.FallbackMail{Domain: "", ReplyTo: ""}
	_, err = e.SendMail("", "contact@wayplot.com", "", util.ConvertSliceTypeRev(toField), subject, body)
	if err != nil {
		log.Println("***>***> EventUserSignedUp: unexpected/unhandled error occurred when user signedup. error:", err)
	}

	return nil
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

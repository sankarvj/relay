package job

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"go.opencensus.io/trace"
)

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

func activityEventEntity(ctx context.Context, accountID, baseEntityID string, db *sqlx.DB) *string {
	be, err := entity.Retrieve(ctx, accountID, baseEntityID, db)
	if err != nil {
		return nil
	}
	props := be.Props()
	for _, prop := range props {
		if prop.Name == "default-activity" {
			return &prop.RefID
		}
	}
	return nil
}

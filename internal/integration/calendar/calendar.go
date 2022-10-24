package calendar

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/discovery"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	intg "gitlab.com/vjsideprojects/relay/internal/integration"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration/calendar"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
)

var (

	// ErrIntegNotFound is used when a specific integrations is requested but none/more than one exist at a time.
	ErrIntegNotFound = errors.New("Integrations not found")
)

type Calendar struct {
}

func (c Calendar) Act(ctx context.Context, accountID string, actionID string, actionPayload intg.ActionPayload, db *sqlx.DB) error {
	teamID := "0" //since the calendar StateAccountLevel is 1 this will work without teamID
	calConfigItem, updaterFunc, err := calendarEntityItem(ctx, accountID, teamID, db)
	if err != nil {
		return err
	}

	switch actionID {
	case "SYNC":
		c := calendar.Gcalendar{OAuthFile: "config/dev/google-apps-client-secret.json", TokenJson: calConfigItem.APIKey}
		syncToken, err := c.Sync(calConfigItem.ID, calConfigItem.SyncToken)
		if err != nil {
			return err
		}
		calConfigItem.SyncToken = syncToken
		updaterFunc(ctx, calConfigItem, db)
	default:
	}
	return nil
}

func CreateCalendarEvent(ctx context.Context, accountID, teamID, entityID, itemID string, valueAddedCalendarFields []entity.Field, db *sqlx.DB) error {
	namedFieldsObj := entity.NamedFieldsObjMap(valueAddedCalendarFields)
	meetingID := uuid.New().String()
	meeting := &integration.Meeting{
		ID:          meetingID,
		Summary:     entityFieldVal(namedFieldsObj["cal_title"]),
		Description: entityFieldVal(namedFieldsObj["summary"]),
		StartTime:   entityFieldVal(namedFieldsObj["start_time"]),
		EndTime:     entityFieldVal(namedFieldsObj["end_time"]),
		Attendees:   namedFieldsObj["attendess"].ChoicesValues(),
	}

	st, err := util.ParseTime(meeting.StartTime)
	if err != nil {
		return err
	}
	end, err := util.ParseTime(meeting.EndTime)
	if err != nil {
		return err
	}

	meeting.StartTime = util.FormatTimeGoogle(st)
	meeting.EndTime = util.FormatTimeGoogle(end)

	calConfigItem, _, err := calendarEntityItem(ctx, accountID, teamID, db)
	if err != nil {
		return err
	}

	c := calendar.Gcalendar{OAuthFile: "config/dev/google-apps-client-secret.json", TokenJson: calConfigItem.APIKey}

	err = c.EventCreate(calConfigItem.ID, meeting)
	if err != nil {
		return err
	}

	ns := discovery.NewDiscovery{
		ID:        meetingID,
		AccountID: accountID,
		EntityID:  entityID,
		ItemID:    itemID,
	}

	_, err = discovery.Create(ctx, db, ns, time.Now())
	if err != nil {
		return err
	}
	return nil
}

func calendarConfigItem(ctx context.Context, discovery discovery.Discover, db *sqlx.DB, sdb *database.SecDB) (entity.CaldendarEntity, entity.UpdaterFunc, error) {
	var calendarEntity entity.CaldendarEntity
	valueAddedFields, updateFunc, err := entity.RetrieveFixedItem(ctx, discovery.AccountID, discovery.EntityID, discovery.ItemID, db, sdb)
	if err != nil {
		return calendarEntity, updateFunc, err
	}
	err = entity.ParseFixedEntity(valueAddedFields, &calendarEntity)
	if err != nil {
		return calendarEntity, updateFunc, err
	}
	return calendarEntity, updateFunc, nil
}

func calendarEntityItem(ctx context.Context, accountID, teamID string, db *sqlx.DB) (entity.CaldendarEntity, entity.UpdaterFunc, error) {
	var calendarEntityItem entity.CaldendarEntity
	valueAddedFields, updateFunc, err := entity.RetriveFixedItemByCategory(ctx, accountID, teamID, entity.FixedEntityCalendar, db)
	if err != nil {
		return calendarEntityItem, nil, err
	}
	err = entity.ParseFixedEntity(valueAddedFields, &calendarEntityItem)
	if err != nil {
		return calendarEntityItem, nil, err
	}
	return calendarEntityItem, updateFunc, nil
}

func entityFieldVal(f entity.Field) string {
	if f.Value != nil {
		return f.Value.(string)
	}
	return ""
}

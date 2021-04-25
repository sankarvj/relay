package calendar

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration"
)

var (
	// ErrIntegNotFound is used when a specific integrations is requested but none/more than one exist at a time.
	ErrIntegNotFound = errors.New("Integrations not found")
)

func CreateEvent(ctx context.Context, accountID string, actionPayload integration.ActionPayload, db *sqlx.DB) error {

	calendarEntity, err := entity.RetrieveFixedEntity(ctx, db, accountID, entity.FixedEntityCalendar)
	if err != nil {
		return err
	}

	items, err := item.EntityItems(ctx, calendarEntity.ID, db)
	if err != nil {
		return err
	}

	if len(items) != 1 {
		return ErrIntegNotFound
	}
	item := items[0]
	valueAddedFields := entity.ValueAddFields(calendarEntity.FieldsIgnoreError(), item.Fields())
	var calendarEntityItem entity.EmailConfigEntity
	err = entity.ParseFixedEntity(valueAddedFields, &calendarEntityItem)
	if err != nil {
		return err
	}

	integration.CreateEvent("config/dev/google-apps-client-secret.json", calendarEntityItem.APIKey)
	return nil
}

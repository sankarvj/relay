package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"go.opencensus.io/trace"
)

// Item represents the Item API method handler set.
type Item struct {
	db            *sqlx.DB
	authenticator *auth.Authenticator
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// List returns all the existing entities associated with team
func (i *Item) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Item.List")
	defer span.End()

	e, fields, err := entity.RetrieveWithFields(ctx, i.db, params["entity_id"])
	if err != nil {
		return err
	}

	items, err := item.List(ctx, e.ID, i.db)
	if err != nil {
		return err
	}

	viewModelItems := make([]item.ViewModelItem, len(items))
	for i, item := range items {
		viewModelItems[i] = createViewModelItem(item)
	}

	response := struct {
		Items    []item.ViewModelItem `json:"items"`
		Category int                  `json:"category"`
		Fields   []entity.Field       `json:"fields"`
	}{
		Items:    viewModelItems,
		Category: e.Category,
		Fields:   fields,
	}
	return web.Respond(ctx, w, response, http.StatusOK)
}

// TimeSeriesList returns all the existing items associated with the entity
func (i *Item) TimeSeriesList(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Item.TimeSeriesList")
	defer span.End()

	e, fields, err := entity.RetrieveWithFields(ctx, i.db, params["entity_id"])
	if err != nil {
		return err
	}

	items, err := item.TimeSeriesList(ctx, e.ID, i.db)
	if err != nil {
		return err
	}

	start := time.Now()
	startRounded := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	itemsMap := item.TimeSeriesSameDayViewModel(items, startRounded, 24)

	response := struct {
		ItemsMap map[time.Time]item.TimeSeriesItem `json:"items_map"`
		Category int                               `json:"category"`
		Fields   []entity.Field                    `json:"fields"`
	}{
		ItemsMap: itemsMap,
		Category: e.Category,
		Fields:   fields,
	}
	return web.Respond(ctx, w, response, http.StatusOK)
}

// Create inserts a new team into the system.
func (i *Item) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Item.Create")
	defer span.End()

	entityID, err := findPrimaryEntityID(ctx, params["team_id"], params["entity_id"], i.db)
	if err != nil {
		return errors.Wrap(err, "")
	}

	var ni item.NewItem
	if err := web.Decode(r, &ni); err != nil {
		return errors.Wrap(err, "")
	}

	item, err := item.Create(ctx, i.db, entityID, ni, time.Now())
	if err != nil {
		return errors.Wrapf(err, "Item: %+v", &item)
	}

	return web.Respond(ctx, w, item, http.StatusCreated)
}

func createViewModelItem(i item.Item) item.ViewModelItem {
	var fields map[string]interface{}
	if err := json.Unmarshal([]byte(i.Input), &fields); err != nil {
		log.Printf("error while unmarshalling item input %v", i.ID)
		log.Println(err)
	}
	return item.ViewModelItem{
		ID:     i.ID,
		Fields: fields,
	}
}

func findPrimaryEntityID(ctx context.Context, teamID, entityID string, db *sqlx.DB) (string, error) {
	if entityID == "0" {
		primaryEntity, err := entity.Primary(ctx, teamID, db)
		if err != nil {
			return entityID, err
		}
		entityID = primaryEntity.ID
	}
	return entityID, nil
}

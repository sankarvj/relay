package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
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
	rPool         *redis.Pool
	db            *sqlx.DB
	authenticator *auth.Authenticator
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// List returns all the existing entities associated with team
func (i *Item) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Item.List")
	defer span.End()

	e, err := entity.Retrieve(ctx, params["entity_id"], i.db)
	if err != nil {
		return err
	}

	fields, err := e.Fields()
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

	e, err := entity.Retrieve(ctx, params["entity_id"], i.db)
	if err != nil {
		return err
	}
	fields, err := e.Fields()
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

	var ni item.NewItem
	if err := web.Decode(r, &ni); err != nil {
		return errors.Wrap(err, "")
	}

	ni.AccountID = params["account_id"]
	ni.EntityID = params["entity_id"]

	ri, err := item.Create(ctx, i.db, ni, time.Now())
	if err != nil {
		return errors.Wrapf(err, "Item: %+v", &i)
	}

	return web.Respond(ctx, w, ri, http.StatusCreated)
}

func createViewModelItem(i item.Item) item.ViewModelItem {
	return item.ViewModelItem{
		ID:     i.ID,
		Fields: i.Fields(),
	}
}

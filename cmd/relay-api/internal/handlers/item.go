package handlers

import (
	"context"
	"net/http"
	"time"

	"gitlab.com/vjsideprojects/relay/internal/relationship"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/reference"
	"go.opencensus.io/trace"
)

// Item represents the Item API method handler set.
type Item struct {
	db            *sqlx.DB
	rPool         *redis.Pool
	authenticator *auth.Authenticator
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// List returns all the existing entities associated with team
//TODO: add pagination
func (i *Item) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Item.List")
	defer span.End()

	e, err := entity.Retrieve(ctx, params["entity_id"], i.db)
	if err != nil {
		return err
	}

	fields, err := e.FieldsWithReference()
	if err != nil {
		return err
	}

	items, err := item.List(ctx, e.ID, i.db)
	if err != nil {
		return err
	}

	viewModelItems := make([]*item.ViewModelItem, len(items))
	for i, item := range items {
		viewModelItem := createViewModelItem(item)
		viewModelItems[i] = &viewModelItem
	}

	reference.UpdateReferenceFields(ctx, fields, viewModelItems, i.db)

	response := struct {
		Items    []*item.ViewModelItem  `json:"items"`
		Category int                    `json:"category"`
		Fields   []*entity.Field        `json:"fields"`
		Entity   entity.ViewModelEntity `json:"entity"`
	}{
		Items:    viewModelItems,
		Category: e.Category,
		Fields:   fields,
		Entity:   createViewModelEntity(e),
	}
	return web.Respond(ctx, w, response, http.StatusOK)
}

// Search returns the items for the given term & key
func (i *Item) Search(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Item.Search")
	defer span.End()

	key := r.URL.Query().Get("k")
	term := r.URL.Query().Get("t")

	e, err := entity.Retrieve(ctx, params["entity_id"], i.db)
	if err != nil {
		return err
	}

	items, err := item.SearchByKey(ctx, e.ID, key, term, i.db)
	if err != nil {
		return err
	}

	viewModelItems := make([]*item.ViewModelItem, len(items))
	for i, item := range items {
		viewModelItem := createViewModelItem(item)
		viewModelItems[i] = &viewModelItem
	}

	response := struct {
		Items []*item.ViewModelItem `json:"items"`
		Key   string                `json:"key"`
	}{
		Items: viewModelItems,
		Key:   key,
	}
	return web.Respond(ctx, w, response, http.StatusOK)
}

//Update updates the item
func (i *Item) Update(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Item.Update")
	defer span.End()

	var vi item.ViewModelItem
	if err := web.Decode(r, &vi); err != nil {
		return errors.Wrap(err, "")
	}
	entityID := params["item_id"]
	existingItem, err := item.Retrieve(ctx, entityID, vi.ID, i.db)
	if err != nil {
		return errors.Wrapf(err, "Item Get During Update")
	}

	err = item.UpdateFields(ctx, i.db, entityID, params["item_id"], vi.Fields)
	if err != nil {
		return errors.Wrapf(err, "Item Update: %+v", &vi)
	}
	//TODO push this to stream/queue
	job.OnFieldUpdate(params["account_id"], params["entity_id"], vi.ID, existingItem.Fields(), vi.Fields, i.db)

	return web.Respond(ctx, w, vi, http.StatusOK)
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
	ni.ID = uuid.New().String()

	ri, err := item.Create(ctx, i.db, ni, time.Now())
	if err != nil {
		return errors.Wrapf(err, "Item: %+v", &i)
	}

	//TODO push this to stream/queue
	job.OnFieldCreate(params["account_id"], params["entity_id"], ni.ID, ni.Fields, i.db)

	return web.Respond(ctx, w, ri, http.StatusCreated)
}

// Retrieve gets the specified item with field meta from the database.
func (i *Item) Retrieve(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Item.Retrieve")
	defer span.End()

	e, err := entity.Retrieve(ctx, params["entity_id"], i.db)
	if err != nil {
		return err
	}

	fields, err := e.FieldsWithReference()
	if err != nil {
		return err
	}

	it, err := item.Retrieve(ctx, params["entity_id"], params["item_id"], i.db)
	if err != nil {
		return err
	}

	bonds, err := relationship.List(ctx, i.db, it.AccountID, it.EntityID)
	if err != nil {
		return err
	}

	viewModelItem := createViewModelItem(it)
	reference.UpdateReferenceFields(ctx, fields, []*item.ViewModelItem{&viewModelItem}, i.db)

	itemDetail := item.ItemDetail{
		Entity: createViewModelEntity(e),
		Item:   viewModelItem,
		Bonds:  bonds,
		Fields: fields,
	}

	return web.Respond(ctx, w, itemDetail, http.StatusOK)
}

func createViewModelItem(i item.Item) item.ViewModelItem {
	return item.ViewModelItem{
		ID:     i.ID,
		Fields: i.Fields(),
	}
}

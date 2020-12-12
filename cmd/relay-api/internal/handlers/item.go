package handlers

import (
	"context"
	"log"
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

	viewModelItems = updateReferenceFields(ctx, fields, viewModelItems, i.db)

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

//Update updates the item
func (i *Item) Update(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Item.Update")
	defer span.End()

	var vi item.ViewModelItem
	if err := web.Decode(r, &vi); err != nil {
		return errors.Wrap(err, "")
	}

	existingItem, err := item.Retrieve(ctx, vi.ID, i.db)
	if err != nil {
		return errors.Wrapf(err, "Item Get During Update")
	}

	err = item.UpdateFields(ctx, i.db, params["item_id"], vi.Fields)
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

	fields, err := e.Fields()
	if err != nil {
		return err
	}

	it, err := item.Retrieve(ctx, params["item_id"], i.db)
	if err != nil {
		return err
	}

	bonds, err := relationship.List(ctx, i.db, it.AccountID, it.EntityID)
	if err != nil {
		return err
	}

	vmItem := createViewModelItem(it)
	viewModelItems := updateReferenceFields(ctx, fields, []item.ViewModelItem{vmItem}, i.db)

	itemDetail := item.ItemDetail{
		Item:   viewModelItems[0],
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

func updateReferenceFields(ctx context.Context, fields []entity.Field, items []item.ViewModelItem, db *sqlx.DB) []item.ViewModelItem {
	refIds := make(map[string]entity.Field, 0) //key as field key and vals are ids
	for _, f := range fields {
		if f.IsReference() {
			f.Value = []interface{}{}
			refIds[f.Key] = f
		}
	}

	//populating map with only ref fieldID
	for _, item := range items {
		for key, val := range item.Fields {
			if f, ok := refIds[key]; ok {
				f.Value = append(f.Value.([]interface{}), val.([]interface{})...)
				refIds[key] = f
			}
		}
	}

	itemFieldsMap := make(map[string]map[string]interface{}, 0)
	for _, f := range refIds {
		refItems, err := item.BulkRetrieve(ctx, f.RefID, removeDuplicateValues(f.Value.([]interface{})), db)
		if err != nil {
			log.Println("error on retriving reference items. Continuing... ", err)
		}
		for _, refIt := range refItems {
			itemFieldsMap[refIt.ID] = refIt.Fields()
		}
	}

	//repopulating with actual values
	for _, item := range items {
		for key, vals := range item.Fields {
			if f, ok := refIds[key]; ok {
				item.Fields[key] = []interface{}{}
				for _, itemID := range vals.([]interface{}) {
					if fieldMap, ok := itemFieldsMap[itemID.(string)]; ok {
						item.Fields[key] = append(item.Fields[key].([]interface{}), fieldMap[f.DisplayGex()])
					}
				}
			}
		}
	}

	return items

}

func removeDuplicateValues(intSlice []interface{}) []interface{} {
	keys := make(map[interface{}]bool)
	list := []interface{}{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"gitlab.com/vjsideprojects/relay/internal/relationship"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"gitlab.com/vjsideprojects/relay/internal/user"

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

	e, err := entity.Retrieve(ctx, params["account_id"], params["entity_id"], i.db)
	if err != nil {
		return err
	}

	fields, err := e.FilteredFields()
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

	reference.UpdateReferenceFields(ctx, params["account_id"], fields, viewModelItems, map[string]interface{}{}, i.db)

	response := struct {
		Items    []item.ViewModelItem   `json:"items"`
		Category int                    `json:"category"`
		Fields   []entity.Field         `json:"fields"`
		Entity   entity.ViewModelEntity `json:"entity"`
	}{
		Items:    viewModelItems,
		Category: e.Category,
		Fields:   fields,
		Entity:   createViewModelEntity(e),
	}

	return web.Respond(ctx, w, response, http.StatusOK)
}

func (i *Item) Templates(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Item.List")
	defer span.End()

	e, err := entity.Retrieve(ctx, params["account_id"], params["entity_id"], i.db)
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

	reference.UpdateChoicesWrapper(ctx, i.db, params["account_id"], fields)

	response := struct {
		Items    []item.ViewModelItem   `json:"items"`
		Category int                    `json:"category"`
		Fields   []entity.Field         `json:"fields"`
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
	parentEntityID := r.URL.Query().Get("e")
	filterID := r.URL.Query().Get("fi")
	filterKey := r.URL.Query().Get("fk")
	log.Println("filterKey--> ", filterKey)

	e, err := entity.Retrieve(ctx, params["account_id"], params["entity_id"], i.db)
	if err != nil {
		return err
	}
	choices := make([]entity.Choice, 0)
	// Its a fixed wrapper entity. Call the respective items
	if e.Category == entity.CategoryFlow { // temp flow handler
		flows, err := flow.SearchByKey(ctx, params["account_id"], parentEntityID, key, term, i.db)
		if err != nil {
			return err
		}
		for _, flow := range flows {
			choice := entity.Choice{
				ID:           flow.ID,
				DisplayValue: flow.Name,
			}
			choices = append(choices, choice)
		}
	} else if e.Category == entity.CategoryNode { // temp flow handler
		//here filterID is the flowID...
		flows, err := node.SearchByKey(ctx, params["account_id"], filterID, key, term, i.db)
		if err != nil {
			return err
		}
		for _, flow := range flows {
			choice := entity.Choice{
				ID:           flow.ID,
				DisplayValue: flow.Name,
			}
			choices = append(choices, choice)
		}
	} else {
		items, err := item.SearchByKey(ctx, e.ID, key, term, i.db)
		if err != nil {
			return err
		}
		for _, item := range items {
			choice := entity.Choice{
				ID:           item.ID,
				DisplayValue: item.Fields()[key],
			}
			choices = append(choices, choice)
		}
	}

	return web.Respond(ctx, w, choices, http.StatusOK)
}

//Update updates the item
func (i *Item) Update(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Item.Update")
	defer span.End()

	var ni item.NewItem
	if err := web.Decode(r, &ni); err != nil {
		return errors.Wrap(err, "")
	}
	entityID := params["entity_id"]
	existingItem, err := item.Retrieve(ctx, entityID, ni.ID, i.db)
	if err != nil {
		return errors.Wrapf(err, "Item Get During Update")
	}

	it, err := item.UpdateFields(ctx, i.db, entityID, params["item_id"], ni.Fields)
	if err != nil {
		return errors.Wrapf(err, "Item Update: %+v", &ni)
	}
	//TODO push this to stream/queue
	job.EventItemUpdated(params["account_id"], params["entity_id"], ni.ID, ni.Fields, existingItem.Fields(), i.db)

	return web.Respond(ctx, w, createViewModelItem(it), http.StatusOK)
}

// Create inserts a new team into the system.
func (i *Item) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Item.Create")
	defer span.End()

	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return err
	}

	var ni item.NewItem
	if err := web.Decode(r, &ni); err != nil {
		return errors.Wrap(err, "")
	}

	log.Printf("ni %v", ni)
	ni.AccountID = params["account_id"]
	ni.EntityID = params["entity_id"]
	ni.UserID = &currentUserID
	ni.ID = uuid.New().String()

	it, err := item.Create(ctx, i.db, ni, time.Now())
	if err != nil {
		return errors.Wrapf(err, "Item: %+v", &i)
	}

	//TODO push this to stream/queue
	job.EventItemCreated(params["account_id"], params["entity_id"], ni, i.db)

	return web.Respond(ctx, w, createViewModelItem(it), http.StatusCreated)
}

// Retrieve gets the specified item with field meta from the database.
func (i *Item) Retrieve(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Item.Retrieve")
	defer span.End()

	e, err := entity.Retrieve(ctx, params["account_id"], params["entity_id"], i.db)
	if err != nil {
		return err
	}

	fields, err := e.FilteredFields()
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
	reference.UpdateReferenceFields(ctx, params["account_id"], fields, []item.ViewModelItem{viewModelItem}, map[string]interface{}{}, i.db)

	itemDetail := struct {
		Entity entity.ViewModelEntity `json:"entity"`
		Item   item.ViewModelItem     `json:"item"`
		Bonds  []relationship.Bond    `json:"bonds"`
		Fields []entity.Field         `json:"fields"`
	}{
		createViewModelEntity(e),
		viewModelItem,
		bonds,
		fields,
	}
	return web.Respond(ctx, w, itemDetail, http.StatusOK)
}

func createViewModelItem(i item.Item) item.ViewModelItem {
	return item.ViewModelItem{
		ID:     i.ID,
		Name:   i.Name,
		Type:   i.Type,
		State:  i.State,
		Fields: i.Fields(),
	}
}

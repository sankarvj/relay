package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"gitlab.com/vjsideprojects/relay/internal/relationship"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"gitlab.com/vjsideprojects/relay/internal/schema"
	"gitlab.com/vjsideprojects/relay/internal/tests"
	"gitlab.com/vjsideprojects/relay/internal/user"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/ruler"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
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

// List returns all the existing items associated with entity
//TODO: add pagination
func (i *Item) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Item.List")
	defer span.End()

	accountID, entityID, _ := takeAEI(ctx, params, i.db)
	state := util.ConvertStrToInt(r.URL.Query().Get("state"))
	viewID := r.URL.Query().Get("view_id")

	e, err := entity.Retrieve(ctx, accountID, entityID, i.db)
	if err != nil {
		return err
	}

	var items []item.Item
	if viewID == "" {
		var err error
		items, err = item.ListFilterByState(ctx, e.ID, state, i.db)
		if err != nil {
			return err
		}

	} else {
		fl, err := flow.Retrieve(ctx, viewID, i.db)
		if err != nil {
			return err
		}
		conditions := job.NewJabEngine().RunExpGrapher(tests.Context(), i.db, i.rPool, accountID, fl.Expression)
		gSegment := graphdb.BuildGNode(accountID, e.ID, false).MakeBaseGNode("", makeConField(conditions))
		result, err := graphdb.GetResult(i.rPool, gSegment)
		if err != nil {
			return err
		}
		items, err = itemsResp(ctx, i.db, accountID, e, result)
		if err != nil {
			return err
		}
	}

	fields, viewModelItems := itemResponse(e, items)
	reference.UpdateReferenceFields(ctx, accountID, entityID, fields, items, map[string]interface{}{}, i.db, job.NewJabEngine())

	response := struct {
		Items    []ViewModelItem        `json:"items"`
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

	accountID, entityID, _ := takeAEI(ctx, params, i.db)
	key := r.URL.Query().Get("k")
	term := r.URL.Query().Get("t")
	filterID := r.URL.Query().Get("fi")
	// filterKey := r.URL.Query().Get("fk")

	e, err := entity.Retrieve(ctx, accountID, entityID, i.db)
	if err != nil {
		return err
	}
	choices := make([]entity.Choice, 0)
	// Its a fixed wrapper entity. Call the respective items
	if e.Category == entity.CategoryFlow { // temp flow handler
		flows, err := flow.SearchByKey(ctx, accountID, filterID, key, term, i.db)
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
		nodes, err := node.SearchByKey(ctx, accountID, filterID, key, term, i.db)
		if err != nil {
			return err
		}
		for _, node := range nodes {
			choice := entity.Choice{
				ID:           node.ID,
				DisplayValue: node.Name,
			}
			choices = append(choices, choice)
		}
	} else {

		items, err := item.SearchByKey(ctx, e.ID, key, term, i.db)
		if err != nil {
			return err
		}
		for _, item := range items {
			displayV := item.Fields()[key]
			if displayV == nil {
				displayV = item.Name
			}
			choice := entity.Choice{
				ID:           item.ID,
				DisplayValue: displayV,
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

	accountID, entityID, itemID := takeAEI(ctx, params, i.db)
	var ni item.NewItem
	if err := web.Decode(r, &ni); err != nil {
		return errors.Wrap(err, "")
	}
	existingItem, err := item.Retrieve(ctx, entityID, ni.ID, i.db)
	if err != nil {
		return errors.Wrapf(err, "Item Get During Update")
	}

	it, err := item.UpdateFields(ctx, i.db, entityID, itemID, ni.Fields)
	if err != nil {
		return errors.Wrapf(err, "Item Update: %+v", &ni)
	}
	//TODO push this to stream/queue
	(&job.Job{}).EventItemUpdated(accountID, entityID, ni.ID, it.Fields(), existingItem.Fields(), i.db, i.rPool)

	return web.Respond(ctx, w, createViewModelItem(it), http.StatusOK)
}

// Create inserts a new team into the system.
func (i *Item) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Item.Create")
	defer span.End()

	accountID, entityID, _ := takeAEI(ctx, params, i.db)
	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return err
	}

	var ni item.NewItem
	if err := web.Decode(r, &ni); err != nil {
		return errors.Wrap(err, "")
	}

	ni.AccountID = accountID
	ni.EntityID = entityID
	ni.UserID = &currentUserID
	ni.ID = uuid.New().String()

	it, err := item.Create(ctx, i.db, ni, time.Now())
	if err != nil {
		return errors.Wrapf(err, "Item: %+v", &i)
	}

	//TODO push this to stream/queue
	(&job.Job{}).EventItemCreated(accountID, entityID, it.ID, ni.Source, i.db, i.rPool)

	return web.Respond(ctx, w, createViewModelItem(it), http.StatusCreated)
}

// Retrieve gets the specified item with field meta from the database.
func (i *Item) Retrieve(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Item.Retrieve")
	defer span.End()

	accountID, entityID, itemID := takeAEI(ctx, params, i.db)
	baseEntityID := r.URL.Query().Get("be")
	baseItemID := r.URL.Query().Get("bi")
	populateBR, _ := strconv.ParseBool(r.URL.Query().Get("bp")) // blue print

	e, err := entity.Retrieve(ctx, accountID, entityID, i.db)
	if err != nil {
		return err
	}

	fields, err := e.FilteredFields()
	if err != nil {
		return err
	}

	it := item.Item{}
	if itemID != "undefined" {
		it, err = item.Retrieve(ctx, entityID, itemID, i.db)
		if err != nil {
			return err
		}

	}

	bonds, err := relationship.List(ctx, i.db, accountID, params["team_id"], entityID)
	if err != nil {
		return err
	}

	fields, viewModelItems := itemResponse(e, []item.Item{it})
	if len(viewModelItems) == 0 {
		viewModelItems = append(viewModelItems, ViewModelItem{})
	}
	reference.UpdateReferenceFields(ctx, accountID, entityID, fields, []item.Item{it}, map[string]interface{}{baseEntityID: baseItemID}, i.db, job.NewJabEngine())

	if populateBR {
		be, err := entity.Retrieve(ctx, accountID, baseEntityID, i.db)
		if err != nil {
			return err
		}
		for i := 0; i < len(fields); i++ {
			if fields[i].IsReference() {
				reference.ChoicesBluePrint(&fields[i], be)
			} else if fields[i].IsNode() {
				reference.ChoicesBluePrint(&fields[i], be)
			}
		}
	}

	itemDetail := struct {
		Entity entity.ViewModelEntity `json:"entity"`
		Item   ViewModelItem          `json:"item"`
		Bonds  []relationship.Bond    `json:"bonds"`
		Fields []entity.Field         `json:"fields"`
	}{
		createViewModelEntity(e),
		viewModelItems[0],
		bonds,
		fields,
	}
	return web.Respond(ctx, w, itemDetail, http.StatusOK)
}

func (i *Item) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	accountID, entityID, itemID := takeAEI(ctx, params, i.db)
	err := item.Delete(ctx, i.db, accountID, entityID, itemID)
	if err != nil {
		return err
	}
	return web.Respond(ctx, w, "SUCCESS", http.StatusAccepted)
}

func createViewModelItem(i item.Item) ViewModelItem {
	return ViewModelItem{
		ID:     i.ID,
		Name:   i.Name,
		Type:   i.Type,
		State:  i.State,
		Fields: i.Fields(),
	}
}

func makeConField(conditions []ruler.Condition) []graphdb.Field {
	conFields := make([]graphdb.Field, 0)
	for _, c := range conditions {
		operator := graphdb.Operator(c.Operator)
		gf := graphdb.Field{
			Expression: operator,
			Key:        c.Key,
			DataType:   graphdb.DType(c.DataType),
			Value:      c.Value,
		}
		conFields = append(conFields, gf)
	}
	return conFields
}

func itemResponse(e entity.Entity, items []item.Item) ([]entity.Field, []ViewModelItem) {
	viewModelItems := make([]ViewModelItem, len(items))
	for i, item := range items {
		viewModelItems[i] = createViewModelItem(item)
	}

	return e.FieldsIgnoreError(), viewModelItems
}

// ViewModelItem represents the view model of item
// (i.e) it has fields instead of attributes
type ViewModelItem struct {
	ID     string                 `json:"id"`
	Name   *string                `json:"name"`
	Type   int                    `json:"type"`
	State  int                    `json:"state"`
	Fields map[string]interface{} `json:"fields"`
}

//AEI accountID, entityID, itemID
// entityID alone has a twist
// Need to bring the same logic in the middleware too.
func takeAEI(ctx context.Context, params map[string]string, db *sqlx.DB) (string, string, string) {
	entityID := params["entity_id"]
	if schema.IsEntitySeeded(entityID) {
		fixedEntity, err := entity.RetrieveFixedEntity(ctx, db, params["account_id"], params["team_id"], entityID)
		if err == nil {
			entityID = fixedEntity.ID
		}
	}
	return params["account_id"], entityID, params["item_id"]
}

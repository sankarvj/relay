package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"gitlab.com/vjsideprojects/relay/internal/relationship"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"gitlab.com/vjsideprojects/relay/internal/schema"
	"gitlab.com/vjsideprojects/relay/internal/user"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
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
	viewID := r.URL.Query().Get("view_id")
	exp := r.URL.Query().Get("exp")
	ls := r.URL.Query().Get("ls")
	sortby := r.URL.Query().Get("sortby")
	direction := r.URL.Query().Get("direction")
	page := util.ConvertStrToInt(r.URL.Query().Get("page"))

	e, err := entity.Retrieve(ctx, accountID, entityID, i.db)
	if err != nil {
		return err
	}
	fields := e.FieldsIgnoreError()

	if viewID != "" {
		fl, err := flow.Retrieve(ctx, viewID, i.db)
		if err != nil {
			return err
		}
		exp = util.AddExpression(exp, fl.Expression)
	}

	if page == 0 {
		ls = setRenderer(ctx, ls, e, i.db)
	}

	var viewModelItems []ViewModelItem
	var countMap map[string]int
	piper := Piper{Viable: e.FlowField() != nil}
	if ls == entity.MetaRenderPipe && page == 0 {
		err := pipeKanban(ctx, accountID, e, &piper, i.db)
		if err != nil {
			return err
		}
		piper.Viable = true
		piper.Pipe = true

		for _, node := range piper.Nodes {
			piper.NodeKey = e.NodeField().Key
			newExp := fmt.Sprintf("{{%s.%s}} eq {%s}", e.ID, e.NodeField().Key, node.ID)
			exp = util.AddExpression(exp, newExp)
			//TODO: segment call doesn't need the count. But it is executing count query in the call. Shall we stop it?
			vitems, _, err := NewSegmenter(exp).
				AddPage(page).
				AddSortLogic(sortby, direction).
				filterWrapper(ctx, accountID, e.ID, fields, map[string]interface{}{}, i.db, i.rPool)
			if err != nil {
				return err
			}
			piper.Items[node.ID] = vitems
		}
	} else {
		viewModelItems, countMap, err = NewSegmenter(exp).
			AddPage(page).
			AddSortLogic(sortby, direction).
			AddCount().
			filterWrapper(ctx, accountID, e.ID, fields, map[string]interface{}{}, i.db, i.rPool)
		if err != nil {
			return err
		}
	}

	//NOT SO GOOD - DOING THIS IS TO SHOW THE LIST OF ENTITIES IN THE ITEMS LIST PAGE FOR UI NAVIGATION
	entities, err := entity.List(ctx, params["account_id"], params["team_id"], []int{entity.CategoryData}, i.db)
	if err != nil {
		return err
	}

	viewModelEntities := make([]entity.ViewModelEntity, len(entities))
	for i, entt := range entities {
		viewModelEntities[i] = createViewModelEntity(entt)
	}

	response := struct {
		Items    []ViewModelItem          `json:"items"`
		Category int                      `json:"category"`
		Fields   []entity.Field           `json:"fields"`
		Entity   entity.ViewModelEntity   `json:"entity"`
		Piper    Piper                    `json:"piper"`
		CountMap map[string]int           `json:"count_map"`
		Entities []entity.ViewModelEntity `json:"entities"`
	}{
		Items:    viewModelItems,
		Category: e.Category,
		Fields:   fields,
		Entity:   createViewModelEntity(e),
		Piper:    piper,
		CountMap: countMap,
		Entities: viewModelEntities,
	}

	return web.Respond(ctx, w, response, http.StatusOK)
}

func (i *Item) StateRecords(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Item.LimitedList")
	defer span.End()

	accountID, entityID, _ := takeAEI(ctx, params, i.db)
	state := util.ConvertStrToInt(params["state"])

	e, err := entity.Retrieve(ctx, accountID, entityID, i.db)
	if err != nil {
		return err
	}
	fields := e.FieldsIgnoreError()

	items, err := item.ListFilterByState(ctx, entityID, state, i.db)
	if err != nil {
		return err
	}
	viewModelItems := itemResponse(items)

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
	} else if e.Category == entity.CategoryChildUnit || term == "" {
		items, err := item.SearchByKey(ctx, e.ID, key, term, i.db)
		if err != nil {
			return err
		}
		whoMap := e.WhoFields()
		choices = choiceResponse(key, items, whoMap)
	} else {
		exp := fmt.Sprintf("{{%s.%s}} lk {%s}", e.ID, key, term)
		result, _, err := NewSegmenter(exp).
			segment(ctx, accountID, e.ID, i.db, i.rPool)
		if err != nil {
			return err
		}
		whoMap := e.WhoFields()
		items, err := itemsResp(ctx, i.db, accountID, result)
		if err != nil {
			return err
		}
		choices = choiceResponse(key, items, whoMap)

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

	errorMap := validateItemCreate(ctx, accountID, entityID, ni.Fields, i.db, i.rPool)
	if errorMap != nil {
		return web.Respond(ctx, w, errorMap, http.StatusForbidden)
	}

	it, err := createAndPublish(ctx, ni, i.db, i.rPool)
	if err != nil {
		return errors.Wrapf(err, "Item: %+v", &i)
	}

	return web.Respond(ctx, w, createViewModelItem(it), http.StatusCreated)
}

func createAndPublish(ctx context.Context, ni item.NewItem, db *sqlx.DB, rp *redis.Pool) (item.Item, error) {
	it, err := item.Create(ctx, db, ni, time.Now())
	if err != nil {
		return item.Item{}, err
	}
	//TODO push this to stream/queue
	(&job.Job{}).EventItemCreated(ni.AccountID, ni.EntityID, it.ID, ni.Source, db, rp)
	return it, err
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

	fields := e.FieldsIgnoreError()
	viewModelItems := itemResponse([]item.Item{it})
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
	// Delete calls should be called in the last stage of job
	// err := item.Delete(ctx, i.db, accountID, entityID, itemID)
	// if err != nil {
	// 	return err
	// }
	(&job.Job{}).EventItemDeleted(accountID, entityID, itemID, i.db, i.rPool)
	return web.Respond(ctx, w, "SUCCESS", http.StatusAccepted)
}

func createViewModelItem(i item.Item) ViewModelItem {
	return ViewModelItem{
		ID:       i.ID,
		EntityID: i.EntityID,
		StageID:  i.StageID,
		Name:     i.Name,
		Type:     i.Type,
		State:    i.State,
		Fields:   i.Fields(),
	}
}

func itemResponse(items []item.Item) []ViewModelItem {
	viewModelItems := make([]ViewModelItem, len(items))
	for i, item := range items {
		viewModelItems[i] = createViewModelItem(item)
	}
	return viewModelItems
}

func choiceResponse(key string, items []item.Item, whoMap map[string]string) []entity.Choice {
	choices := make([]entity.Choice, 0)
	for _, item := range items {
		//display
		displayV := item.Fields()[key]
		if displayV == nil {
			displayV = item.Name
		}

		//avatar
		var avatar string
		if ava, ok := whoMap[entity.WhoAvatar]; ok {
			if aval, ok := item.Fields()[ava]; ok {
				avatar = aval.(string)
			}

		}

		choice := entity.Choice{
			ID:           item.ID,
			DisplayValue: displayV,
			Avatar:       avatar,
		}
		choices = append(choices, choice)
	}
	return choices
}

// ViewModelItem represents the view model of item
// (i.e) it has fields instead of attributes
type ViewModelItem struct {
	ID       string                 `json:"id"`
	EntityID string                 `json:"entity_id"`
	StageID  *string                `json:"stage_id"`
	Name     *string                `json:"name"`
	Type     int                    `json:"type"`
	State    int                    `json:"state"`
	Fields   map[string]interface{} `json:"fields"`
}

// AEI accountID, entityID, itemID
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

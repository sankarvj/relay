package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"gitlab.com/vjsideprojects/relay/internal/relationship"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/schema"
	"gitlab.com/vjsideprojects/relay/internal/user"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/stream"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/reference"
	"go.opencensus.io/trace"
)

// Item represents the Item API method handler set.
type Item struct {
	db            *sqlx.DB
	chdb          *sql.DB
	sdb           *database.SecDB
	authenticator *auth.Authenticator
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// List returns all the existing items associated with entity
// TODO: add pagination
func (i *Item) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Item.List")
	defer span.End()

	//tracker.TestChan().Log("Item Loaded", "Items loaded")

	accountID, entityID, _ := takeAEI(ctx, params, i.db)
	viewID := r.URL.Query().Get("view_id")
	exp := r.URL.Query().Get("exp")
	ls := r.URL.Query().Get("ls")
	sortby := r.URL.Query().Get("sortby")
	groupby := r.URL.Query().Get("groupby")
	direction := r.URL.Query().Get("direction")
	page := util.ConvertStrToInt(r.URL.Query().Get("page"))

	e, err := entity.Retrieve(ctx, accountID, entityID, i.db, i.sdb)
	if err != nil {
		return err
	}
	fields := e.EasyFieldsByRole(ctx)

	piper := Piper{
		Items: make(map[string][]ViewModelItem, 0),
	}

	if !util.IsEmpty(viewID) {
		exp = ""
		fl, err := flow.Retrieve(ctx, viewID, i.db)
		if err != nil {
			return err
		}
		if fl.Mode == flow.FlowModePipeLine { //case: kanban view
			ff := e.FlowField()
			fl.Expression = fmt.Sprintf("{{%s.%s}} eq {%s}", e.ID, ff.Key, fl.ID)
			piper.Nodes, err = nodeStages(ctx, accountID, fl.ID, i.db)
			if err != nil {
				return err
			}
		} else {
			exp = util.AddExpression(exp, fl.Expression)
		}
	}

	var viewModelItems []ViewModelItem
	var countMap map[string]int

	if groupby != "" && page == 0 {
		err := loadItemsWithGroupByLogic(ctx, accountID, e, exp, sortby, direction, groupby, page, &piper, i.db, i.sdb)
		if err != nil {
			return err
		}
	} else if ls == entity.MetaRenderPipe && page == 0 {
		for _, node := range piper.Nodes {
			piper.NodeKey = e.NodeField().Key
			newExp := fmt.Sprintf("{{%s.%s}} eq {%s}", e.ID, e.NodeField().Key, node.ID)
			finalExp := util.AddExpression(exp, newExp)
			vitems, _, err := NewSegmenter(finalExp).
				AddPage(page).
				AddSortLogic(sortby, direction).
				filterWrapper(ctx, accountID, e.ID, fields, map[string]interface{}{}, i.db, i.sdb)
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
			filterWrapper(ctx, accountID, e.ID, fields, map[string]interface{}{}, i.db, i.sdb)
		if err != nil {
			return err
		}
	}

	response := struct {
		Items    []ViewModelItem        `json:"items"`
		Category int                    `json:"category"`
		Fields   []entity.Field         `json:"fields"`
		Entity   entity.ViewModelEntity `json:"entity"`
		Piper    Piper                  `json:"piper"`
		CountMap map[string]int         `json:"count_map"`
	}{
		Items:    viewModelItems,
		Category: e.Category,
		Fields:   fields,
		Entity:   createViewModelEntity(e),
		Piper:    piper,
		CountMap: countMap,
	}

	return web.Respond(ctx, w, response, http.StatusOK)
}

func (i *Item) StateRecords(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Item.LimitedList")
	defer span.End()

	accountID, entityID, _ := takeAEI(ctx, params, i.db)
	state := util.ConvertStrToInt(params["state"])

	e, err := entity.Retrieve(ctx, accountID, entityID, i.db, i.sdb)
	if err != nil {
		return err
	}
	fields := e.EasyFields()

	items, err := item.ListFilterByState(ctx, accountID, entityID, state, i.db)
	if err != nil {
		return err
	}

	userIDs := make(map[string]bool, 0)
	for _, item := range items {
		userIDs[*item.UserID] = true
	}
	uMap, _ := userMap(ctx, accountID, userIDs, i.db)
	viewModelItems := itemResponse(items, uMap, fields)

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

// Update updates the item
func (i *Item) Update(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Item.Update")
	defer span.End()

	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return err
	}

	accountID, entityID, itemID := takeAEI(ctx, params, i.db)
	var ni item.NewItem
	if err := web.Decode(r, &ni); err != nil {
		return errors.Wrap(err, "")
	}

	//retriving the existing item don't change the order of execution
	existingItem, err := item.Retrieve(ctx, accountID, entityID, itemID, i.db)
	if err != nil {
		return errors.Wrapf(err, "error retriving item")
	}

	err = validateMyselfWithOwner(ctx, accountID, entityID, existingItem, i.db, i.sdb)
	if err != nil {
		return err
	}

	it, err := item.UpdateFieldsWithMeta(ctx, i.db, existingItem, ni.Name, ni.Fields, ni.Meta, ni.IsPublic)
	if err != nil {
		return errors.Wrapf(err, "error when update item fields %+v", &ni)
	}
	if it.State == item.StateDefault { // no need to go to job for web forms now!
		go job.NewJob(i.db, i.sdb, i.authenticator.FireBaseAdminSDK).Stream(stream.NewUpdateItemMessage(ctx, i.db, accountID, currentUserID, entityID, ni.ID, it.Fields(), existingItem.Fields()))
	}
	return web.Respond(ctx, w, createViewModelItem(it, nil), http.StatusOK)
}

// Create inserts a new team into the system.
func (i *Item) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Item.Create")
	defer span.End()

	accountID, entityID, _ := takeAEI(ctx, params, i.db)
	currentUser, err := user.RetrieveCurrentUser(ctx, accountID, i.db)
	if err != nil {
		return err
	}

	var ni item.NewItem
	if err := web.Decode(r, &ni); err != nil {
		return errors.Wrap(err, "")
	}

	ni.AccountID = accountID
	ni.EntityID = entityID
	ni.UserID = &currentUser.ID
	ni.ID = uuid.New().String()
	ni.GenieID = util.PickGenieID(ni.Source)

	errorMap := validateItemCreate(ctx, accountID, entityID, ni.Fields, i.db, i.sdb)
	if errorMap != nil {
		return web.Respond(ctx, w, errorMap, http.StatusForbidden)
	}

	it, err := createAndPublish(ctx, currentUser.ID, ni, i.db, i.sdb, i.authenticator.FireBaseAdminSDK)
	if err != nil {
		return errors.Wrapf(err, "Item: %+v", &i)
	}

	vmItem := createViewModelItem(it, nil)
	vmItem.UserName = *currentUser.Name
	vmItem.UserAvatar = *currentUser.Avatar

	return web.Respond(ctx, w, vmItem, http.StatusCreated)
}

func createAndPublish(ctx context.Context, userID string, ni item.NewItem, db *sqlx.DB, sdb *database.SecDB, fbSDKPath string) (item.Item, error) {
	it, err := item.Create(ctx, db, ni, time.Now())
	if err != nil {
		return item.Item{}, err
	}

	if it.State == item.StateWebForm { // no need to go to job for web forms now!
		return it, err
	}

	//stream
	go job.NewJob(db, sdb, fbSDKPath).Stream(stream.NewCreteItemMessage(ctx, db, ni.AccountID, userID, ni.EntityID, it.ID, ni.Source))
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

	e, err := entity.Retrieve(ctx, accountID, entityID, i.db, i.sdb)
	if err != nil {
		return err
	}
	fields := e.EasyFieldsByRole(ctx)

	it := item.MakeItem(populateBR)
	bonds := []relationship.Bond{}
	if itemID != "undefined" { // get call
		it, err = item.Retrieve(ctx, accountID, entityID, itemID, i.db)
		if err != nil {
			return err
		}
		//TODO not needed to populate this all the time
		bonds, err = relationship.List(ctx, i.db, accountID, params["team_id"], entityID, auth.God(ctx))
		if err != nil {
			return err
		}
	} else { //create call
		//update date and time fields with current time
		for i := 0; i < len(fields); i++ {
			if fields[i].IsDateOrTime() {
				fields[i].Value = time.Now()
			}
		}
	}
	uMap := make(map[string]*user.User, 0)
	if it.UserID != nil {
		uMap, _ = userMap(ctx, accountID, map[string]bool{*it.UserID: true}, i.db)
	}

	viewModelItems := itemResponse([]item.Item{it}, uMap, fields)
	if len(viewModelItems) == 0 {
		viewModelItems = append(viewModelItems, ViewModelItem{})
	}

	viewModelBonds := make([]relationship.Bond, 0)
	for _, b := range bonds {
		if b.Category != entity.CategoryNode {
			viewModelBonds = append(viewModelBonds, b)
		}
	}

	associatedEntities, err := entity.BulkRetrieve(ctx, entityIds(bonds), i.db)
	if err != nil {
		return err
	}

	viewModelEntities := make([]entity.ViewModelEntity, len(associatedEntities))
	for i, entt := range associatedEntities {
		if entt.Category != entity.CategoryNode && entt.Category != entity.CategoryFlow {
			viewModelEntities[i] = createViewModelEntity(entt)
		}
	}

	reference.UpdateReferenceFields(ctx, accountID, entityID, fields, []item.Item{it}, map[string]interface{}{baseEntityID: baseItemID}, i.db, i.sdb, job.NewJabEngine())

	itemDetail := struct {
		Entity        entity.ViewModelEntity   `json:"entity"`
		Item          ViewModelItem            `json:"item"`
		Bonds         []relationship.Bond      `json:"bonds"`
		Fields        []entity.Field           `json:"fields"`
		ChildEntities []entity.ViewModelEntity `json:"child_entities"`
	}{
		createViewModelEntity(e),
		viewModelItems[0],
		viewModelBonds,
		fields,
		viewModelEntities,
	}
	return web.Respond(ctx, w, itemDetail, http.StatusOK)
}

func (i *Item) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	accountID, entityID, itemID := takeAEI(ctx, params, i.db)

	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return err
	}

	err = validateMyself(ctx, accountID, entityID, itemID, i.db)
	if err != nil {
		return err
	}

	// Delete calls should be called in the last stage of job
	// err := item.Delete(ctx, i.db, accountID, entityID, itemID)
	// if err != nil {
	// 	return err
	// }

	//stream
	go job.NewJob(i.db, i.sdb, i.authenticator.FireBaseAdminSDK).Stream(stream.NewDeleteItemMessage(ctx, i.db, accountID, currentUserID, entityID, itemID))

	return web.Respond(ctx, w, "SUCCESS", http.StatusAccepted)
}

func (i *Item) ToggleAccess(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Item.ToggleAccess")
	defer span.End()

	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return err
	}

	accountID, entityID, itemID := takeAEI(ctx, params, i.db)
	if util.NotEmpty(entityID) && util.NotEmpty(itemID) {
		existingItem, err := item.Retrieve(ctx, accountID, entityID, itemID, i.db)
		if err != nil {
			return errors.Wrapf(err, "error retriving item")
		}
		err = validateMyself(ctx, accountID, entityID, itemID, i.db)
		if err != nil {
			return err
		}
		existingItem.IsPublic = !existingItem.IsPublic
		it, err := item.UpdatePublicAccess(ctx, i.db, existingItem)
		if err != nil {
			return errors.Wrapf(err, "error when toggle is_public for item")
		}
		go job.NewJob(i.db, i.sdb, i.authenticator.FireBaseAdminSDK).Stream(stream.NewUpdateItemMessage(ctx, i.db, accountID, currentUserID, entityID, existingItem.ID, it.Fields(), existingItem.Fields()))
		return web.Respond(ctx, w, createViewModelItem(it, nil), http.StatusOK)
	}
	return web.Respond(ctx, w, "failure", http.StatusNotAcceptable)
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

func entityIds(bonds []relationship.Bond) []string {
	ids := make([]string, len(bonds))
	for i, b := range bonds {
		ids[i] = b.EntityID
	}
	return ids
}

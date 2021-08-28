package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
)

// Check provides support for orchestration health checks.
type Event struct {
	db    *sqlx.DB
	rPool *redis.Pool
}

// Create inserts a new team into the system.
func (ev *Event) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Event.Create")
	defer span.End()

	accountID, entityID, _ := takeAEI(ctx, params, ev.db)
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

	it, err := item.Create(ctx, ev.db, ni, time.Now())
	if err != nil {
		return errors.Wrapf(err, "Item: %+v", &it)
	}

	return web.Respond(ctx, w, createViewModelItem(it), http.StatusCreated)
}

func (ev *Event) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Event.List")
	defer span.End()

	accountID, entityID, itemID := takeAEI(ctx, params, ev.db)

	enty, err := entity.Retrieve(ctx, accountID, entityID, ev.db)
	if err != nil {
		return err
	}
	props := enty.Props()
	eventEntityIds := make([]string, 0)
	for _, prop := range props {
		// here add the filter from the UI
		eventEntityIds = append(eventEntityIds, prop.RefID)
	}

	eventEntities, err := entity.BulkRetrieve(ctx, eventEntityIds, ev.db)
	if err != nil {
		return err
	}
	eventEntitieMap := make(map[string]entity.Entity, 0)
	for _, e := range eventEntities {
		eventEntitieMap[e.ID] = e
	}

	items, err := item.GenieEntityItems(ctx, eventEntityIds, itemID, ev.db)
	if err != nil {
		return err
	}

	viewModelEvents := createViewModelEvents(eventEntitieMap, items)

	response := struct {
		Events []ViewModelEvent `json:"events"`
	}{
		Events: viewModelEvents,
	}

	return web.Respond(ctx, w, response, http.StatusOK)
}

func createViewModelEvents(entityMap map[string]entity.Entity, items []item.Item) []ViewModelEvent {
	viewModelEvents := make([]ViewModelEvent, 0)
	for _, it := range items {
		enty := entityMap[it.EntityID]
		fields := enty.FieldsIgnoreError()
		itemFieldsMap := it.Fields()

		dynamicPlaceHolder := make(map[string]interface{}, 0)
		// value add properties
		for i := 0; i < len(fields); i++ {
			if val, ok := itemFieldsMap[fields[i].Key]; ok {
				fields[i].Value = val
				dynamicPlaceHolder[fields[i].Meta["layout"]] = val
			}
		}

		viewModelEvent := ViewModelEvent{
			EventID:         it.ID,
			EventEntity:     it.EntityID,
			EventEntityName: enty.DisplayName,
			UserName:        *it.UserID,
			Action:          dynamicPlaceHolder["action"],
			Title:           dynamicPlaceHolder["title"],
			Footer:          dynamicPlaceHolder["footer"],
		}
		viewModelEvents = append(viewModelEvents, viewModelEvent)
	}
	return viewModelEvents
}

type ViewModelEvent struct {
	EventID         string      `json:"event_id"`
	EventEntity     string      `json:"event_entity"`
	EventEntityName string      `json:"event_entity_name"`
	UserName        string      `json:"user_name"`
	Action          interface{} `json:"action"` //lable:action - created, clicked, viewed, updated, etc
	Title           interface{} `json:"title"`  //lable:title  - task, deal, amazon.com
	Footer          interface{} `json:"footer"` //lable:footer - 8 times
	Time            string      `json:"time"`
}
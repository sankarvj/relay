package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/connection"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/schema"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
)

// Check provides support for orchestration health checks.
type Event struct {
	db *sqlx.DB
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
	sourceItemID := params["item_id"]
	accountID := params["account_id"]
	next := r.URL.Query().Get("next")

	connections, err := connection.JustChildItemIDs(ctx, ev.db, accountID, sourceItemID, next)
	if err != nil {
		return errors.Wrap(err, "selecting related connections")
	}

	if len(connections) == 5 {
		next = connections[len(connections)-1].ConnectionID
	} else {
		next = ""
	}

	uMap, err := userMap(ctx, connections, ev.db)
	if err != nil {
		return errors.Wrap(err, "forming users map")
	}

	viewModelEvents := make([]ViewModelEvent, 0)
	for _, c := range connections {
		avatar, name, email := userAvatarNameEmail(uMap[c.UserID])
		eventID, eventEntityID := c.PickOpposite(sourceItemID)
		viewModelEvent := ViewModelEvent{
			EventID:         eventID,
			EventEntity:     eventEntityID,
			EventEntityName: c.EntityName,
			UserAvatar:      avatar,
			UserName:        name,
			UserEmail:       email,
			Action:          c.Action,
			Title:           c.Title,
			Footer:          c.SubTitle,
			Time:            c.CreatedAt,
		}
		viewModelEvents = append(viewModelEvents, viewModelEvent)
	}

	response := struct {
		Events []ViewModelEvent `json:"events"`
		Next   string           `json:"next"`
	}{
		Events: viewModelEvents,
		Next:   next,
	}

	return web.Respond(ctx, w, response, http.StatusOK)
}

func userMap(ctx context.Context, connections []connection.Connection, db *sqlx.DB) (map[string]*user.User, error) {
	userMap := make(map[string]*user.User, 0)
	userIDs := make(map[string]bool, 0)
	for _, c := range connections {
		userIDs[c.UserID] = true
	}
	users, err := user.BulkRetrieveUsers(ctx, userkeys(userIDs), db)
	if err != nil {
		return userMap, err
	}

	for _, u := range users {
		userMap[u.ID] = &u
	}
	userMap[schema.SeedSystemUserID] = &user.User{
		ID:     schema.SeedSystemUserID,
		Name:   util.String("System"),
		Avatar: util.String("https://randomuser.me/api/portraits/thumb/lego/1.jpg"),
	}
	return userMap, nil
}

type ViewModelEvent struct {
	EventID         string      `json:"event_id"`
	EventEntity     string      `json:"event_entity"`
	EventEntityName string      `json:"event_entity_name"`
	UserAvatar      string      `json:"user_avatar"`
	UserName        string      `json:"user_name"`
	UserEmail       string      `json:"user_email"`
	Action          interface{} `json:"action"` //lable:action - created, clicked, viewed, updated, etc
	Title           interface{} `json:"title"`  //lable:title  - task, deal, amazon.com
	Footer          interface{} `json:"footer"` //lable:footer - 8 times
	Time            time.Time   `json:"time"`
}

func userkeys(oneMap map[string]bool) []string {
	keys := make([]string, 0, len(oneMap))
	for k := range oneMap {
		keys = append(keys, k)
	}
	return keys
}

func userAvatarNameEmail(u *user.User) (string, string, string) {
	if u != nil {
		return *u.Avatar, *u.Name, u.Email
	}
	return "", "", ""
}

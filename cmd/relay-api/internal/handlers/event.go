package handlers

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/connection"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/schema"
	"gitlab.com/vjsideprojects/relay/internal/user"
)

// Check provides support for orchestration health checks.
type Event struct {
	db *sqlx.DB
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

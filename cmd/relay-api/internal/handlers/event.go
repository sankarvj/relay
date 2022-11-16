package handlers

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/connection"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
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

	userIDs := make(map[string]bool, 0)
	for _, c := range connections {
		userIDs[c.UserID] = true
	}
	uMap, err := userMap(ctx, accountID, userIDs, ev.db)
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

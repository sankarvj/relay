package handlers

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/connection"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/relationship"
	"go.opencensus.io/trace"
)

// Relationship represents the Relationship API method handler set.
type Relationship struct {
	db            *sqlx.DB
	authenticator *auth.Authenticator
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// List returns all the existing relationships associated with entity_id
func (rs *Relationship) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Relationship.List")
	defer span.End()

	accounts, err := relationship.List(ctx, rs.db, params["account_id"], params["entity_id"])
	if err != nil {
		return errors.Wrap(err, "selecting relationships for the entity id")
	}

	return web.Respond(ctx, w, accounts, http.StatusOK)
}

func (rs *Relationship) ChildItems(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Connections.List")
	defer span.End()
	actualEntityID := params["entity_id"]
	e, err := entity.Retrieve(ctx, actualEntityID, rs.db)
	if err != nil {
		return err
	}

	fields, err := e.Fields()
	if err != nil {
		return err
	}

	//TODO: add pagination
	itemIDs, err := connection.ChildItemIDs(ctx, rs.db, params["account_id"], params["relationship_id"], params["item_id"])
	if err != nil {
		return errors.Wrap(err, "selecting related item ids")
	}

	childItems, err := item.BulkRetrieve(ctx, actualEntityID, itemIDs, rs.db)
	if err != nil {
		return errors.Wrap(err, "fetching items from selected ids")
	}

	viewModelItems := make([]item.ViewModelItem, len(childItems))
	for i, item := range childItems {
		viewModelItems[i] = createViewModelItem(item)
	}

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

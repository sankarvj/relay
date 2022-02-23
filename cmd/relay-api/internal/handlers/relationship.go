package handlers

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/connection"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/reference"
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

	relationships, err := relationship.List(ctx, rs.db, params["account_id"], params["team_id"], params["entity_id"])
	if err != nil {
		return errors.Wrap(err, "selecting relationships for the entity id")
	}

	return web.Respond(ctx, w, relationships, http.StatusOK)
}

func (rs *Relationship) ChildItems(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Connections.List")
	defer span.End()
	sourceEntityID := params["entity_id"]
	sourceItemID := params["item_id"]
	accountID := params["account_id"]
	relationshipID := params["relationship_id"]
	ls := r.URL.Query().Get("ls")
	page := util.ConvertStrToInt(r.URL.Query().Get("page"))

	relation, err := relationship.Retrieve(ctx, accountID, relationshipID, rs.db)
	if err != nil {
		return err
	}

	relatedEntityID := relation.SrcEntityID
	if relatedEntityID == sourceEntityID {
		relatedEntityID = relation.DstEntityID
	}

	e, err := entity.Retrieve(ctx, accountID, relatedEntityID, rs.db)
	if err != nil {
		return err
	}

	fields, err := e.FilteredFields()
	if err != nil {
		return err
	}

	//There are three ways to fetch the child ids
	// 1. Fetch child item ids by querying the connections table.
	// 2. Fetch child item ids by querying the graph db. tick
	// 3. Fetch child item ids by querying the parent_item_id (formerly genie_id)

	//TODO: add pagination
	itemIDs, err := connection.ChildItemIDs(ctx, rs.db, accountID, relationshipID, sourceItemID)
	if err != nil {
		return errors.Wrap(err, "selecting related item ids")
	}

	piper := Piper{Viable: true}
	if ls == entity.MetaRenderPipe && page == 0 {
		err := pipeKanban(ctx, e, &piper, rs.db)
		if err != nil {
			return err
		}
		piper.Viable = true
		piper.Pipe = true
	}

	childItems, err := item.BulkRetrieve(ctx, e.ID, itemIDs, rs.db)
	if err != nil {
		return errors.Wrap(err, "fetching items from selected ids")
	}

	sourceMap := make(map[string]interface{}, 0)
	sourceMap[sourceEntityID] = sourceItemID
	//When populating the fields for the child items please populate the parent id also
	fields = e.FieldsIgnoreError()
	viewModelItems := itemResponse(childItems)
	reference.UpdateReferenceFields(ctx, accountID, relatedEntityID, fields, childItems, sourceMap, rs.db, job.NewJabEngine())

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

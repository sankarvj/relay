package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"go.opencensus.io/trace"
)

// Entity represents the Entity API method handler set.
type Entity struct {
	db            *sqlx.DB
	authenticator *auth.Authenticator
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// List returns all the existing entities associated with team
func (e *Entity) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Entity.List")
	defer span.End()

	entities, err := entity.List(ctx, params["team_id"], e.db)
	if err != nil {
		return err
	}

	viewModelEntities := make([]entity.ViewModelEntity, len(entities))
	for i, entity := range entities {
		viewModelEntities[i] = createViewModelEntity(entity)
	}

	return web.Respond(ctx, w, viewModelEntities, http.StatusOK)
}

// Create inserts a new team into the system.
func (e *Entity) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Entity.Create")
	defer span.End()

	var ne entity.NewEntity
	if err := web.Decode(r, &ne); err != nil {
		return errors.Wrap(err, "")
	}

	//set account_id from the request path
	ne.TeamID = params["team_id"]
	entity, err := entity.Create(ctx, e.db, ne, time.Now())
	if err != nil {
		return errors.Wrapf(err, "Entity: %+v", &entity)
	}

	return web.Respond(ctx, w, entity, http.StatusCreated)
}

func createViewModelEntity(e entity.Entity) entity.ViewModelEntity {
	var fields []entity.Field
	if err := json.Unmarshal([]byte(e.Attributes), &fields); err != nil {
		log.Printf("error while unmarshalling entity attributes %v", e.ID)
		log.Println(err)
	}

	return entity.ViewModelEntity{
		ID:          e.ID,
		TeamID:      e.TeamID,
		Name:        e.Name,
		Description: e.Description,
		AssigneeID:  e.AssigneeID,
		State:       e.State,
		Mode:        e.Mode,
		Priority:    e.Priority,
		Retry:       e.Retry,
		Fields:      fields,
		Tags:        e.Tags,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

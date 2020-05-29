package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/rule"
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

// Retrieve returns the specified entity from the system.
func (e *Entity) Retrieve(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Entity.Retrieve")
	defer span.End()

	entity, err := entity.Retrieve(ctx, params["entity_id"], e.db)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, createViewModelEntity(*entity), http.StatusOK)
}

// Create inserts a new team into the system.
func (e *Entity) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Entity.Create")
	defer span.End()

	var ne entity.NewEntity
	if err := web.Decode(r, &ne); err != nil {
		return errors.Wrap(err, "")
	}
	//add key with a UUID
	fieldKeyBinder(ne.Fields)

	teamID, _ := strconv.ParseInt(params["team_id"], 10, 64)
	//set account_id from the request path
	entity, err := entity.Create(ctx, e.db, teamID, ne, time.Now())
	if err != nil {
		return errors.Wrapf(err, "Entity: %+v", &entity)
	}

	return web.Respond(ctx, w, entity, http.StatusCreated)
}

//Trigger triggers the entity to start the flow of evaluation of rules
func (e *Entity) Trigger(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Entity.Trigger")
	defer span.End()

	entity, err := entity.Retrieve(ctx, params["entity_id"], e.db)
	if err != nil {
		return err
	}

	rules, err := rule.List(ctx, params["entity_id"], e.db)
	if err != nil {
		return err
	}

	for i := 0; i < len(rules); i++ {
		expression := rules[i].Expression
		rule.RunRuleEngine(ctx, e.db, expression)
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
		Category:    e.Category,
		State:       e.State,
		Mode:        e.Mode,
		Retry:       e.Retry,
		Fields:      fields,
		Tags:        e.Tags,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func fieldKeyBinder(fields []entity.Field) []entity.Field {
	for i := 0; i < len(fields); i++ {
		fields[i].Key = uuid.New().String()
	}
	return fields
}
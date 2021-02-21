package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"go.opencensus.io/trace"
)

// Entity represents the Entity API method handler set.
type Entity struct {
	db            *sqlx.DB
	rPool         *redis.Pool
	authenticator *auth.Authenticator
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// List returns all the existing entities associated with team
func (e *Entity) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Entity.List")
	defer span.End()

	entities, err := entity.List(ctx, params["team_id"], categories(r.URL.Query().Get("category_id")), e.db)
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

	entity, err := entity.Retrieve(ctx, params["account_id"], params["entity_id"], e.db)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, createViewModelEntity(entity), http.StatusOK)
}

// Create inserts a new team into the system.
func (e *Entity) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Entity.Create")
	defer span.End()

	var ne entity.NewEntity
	if err := web.Decode(r, &ne); err != nil {
		return errors.Wrap(err, "")
	}
	ne.ID = uuid.New().String()
	ne.AccountID = params["account_id"]
	ne.TeamID = params["team_id"]
	//add key with a UUID
	fieldKeyBinder(ne.ID, ne.Fields)

	entity, err := entity.Create(ctx, e.db, ne, time.Now())
	if err != nil {
		return errors.Wrapf(err, "Entity: %+v", &entity)
	}

	return web.Respond(ctx, w, entity, http.StatusCreated)
}

//Update updates the entity
func (e *Entity) Update(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Entity.Update")
	defer span.End()

	var ve entity.ViewModelEntity
	if err := web.Decode(r, &ve); err != nil {
		return errors.Wrap(err, "")
	}

	input, err := json.Marshal(ve.Fields)
	if err != nil {
		return errors.Wrap(err, "encode fields to input")
	}

	err = entity.Update(ctx, e.db, params["account_id"], params["entity_id"], string(input), time.Now())
	if err != nil {
		return errors.Wrapf(err, "Entity: %+v", &ve)
	}

	return web.Respond(ctx, w, ve, http.StatusOK)
}

func createViewModelEntity(e entity.Entity) entity.ViewModelEntity {
	fields, err := e.Fields()
	if err != nil {
		log.Println(err)
	}

	return entity.ViewModelEntity{
		ID:          e.ID,
		Name:        e.Name,
		DisplayName: e.DisplayName,
		Category:    e.Category,
		State:       e.State,
		Fields:      fields,
		Tags:        e.Tags,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func fieldKeyBinder(srcEntityId string, fields []entity.Field) []entity.Field {
	nameChecker := make(map[string]bool, len(fields))
	for i := 0; i < len(fields); i++ {
		api_name := fields[i].Name
		if api_name == "" {
			api_name = fields[i].DisplayName
			if _, ok := nameChecker[api_name]; ok {
				api_name = fmt.Sprintf("%s_%s", api_name, util.RandString(5))
			}
			nameChecker[api_name] = true
			fields[i].Name = strings.ReplaceAll(strings.ToLower(api_name), " ", "_")
		}
		fields[i].Key = uuid.New().String()
	}
	return fields
}

func categories(categoryID string) []int {
	ids := []int{}
	i, err := strconv.Atoi(categoryID)
	if err != nil {
		log.Printf("cannot parse category_id from the request %s", err)
		return ids
	} else if i == -1 {
		log.Printf("fetch all categories")
		return ids
	}

	ids = append(ids, i)
	return ids
}

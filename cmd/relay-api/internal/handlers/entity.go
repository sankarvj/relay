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

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/account"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/team"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
)

// Entity represents the Entity API method handler set.
type Entity struct {
	db            *sqlx.DB
	authenticator *auth.Authenticator
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

func (e *Entity) Home(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Entity.Home")
	defer span.End()

	acc, err := account.Retrieve(ctx, e.db, params["account_id"])
	if err != nil {
		return err
	}

	cu, err := user.RetrieveCurrentUser(ctx, e.db)
	if err != nil {
		return err
	}

	cus, err := user.UserSettingRetrieve(ctx, params["account_id"], cu.ID, e.db)
	if err != nil {
		return err
	}

	teams, teamID, err := selectedTeam(ctx, params["account_id"], params["team_id"], e.db)
	if err != nil {
		return err
	}

	entities, err := entity.AccountCoreEntities(ctx, acc.ID, e.db)
	if err != nil {
		return err
	}

	viewModelEntities := make([]entity.ViewModelEntity, len(entities))
	for i, entt := range entities {
		viewModelEntities[i] = createViewModelEntity(entt)
	}

	homeDetail := struct {
		AccountName    string                    `json:"account_name"`
		SelectedTeamID string                    `json:"selected_product_id"`
		Teams          []team.Team               `json:"products"`
		Entities       []entity.ViewModelEntity  `json:"entities"`
		User           user.ViewModelUser        `json:"user"`
		UserSetting    user.ViewModelUserSetting `json:"user_setting"`
	}{
		acc.Name,
		teamID,
		teams,
		viewModelEntities,
		createViewModelUser(*cu),
		createViewModelUS(cus),
	}

	return web.Respond(ctx, w, homeDetail, http.StatusOK)
}

// List returns all the existing entities associated with team
func (e *Entity) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Entity.List")
	defer span.End()

	entities, err := entity.List(ctx, params["account_id"], params["team_id"], categories(r.URL.Query().Get("category_id")), e.db)
	if err != nil {
		return err
	}

	viewModelEntities := make([]entity.ViewModelEntity, len(entities))
	for i, entt := range entities {
		viewModelEntities[i] = createViewModelEntity(entt)
	}

	return web.Respond(ctx, w, viewModelEntities, http.StatusOK)
}

// Retrieve returns the specified entity from the system.
func (e *Entity) Retrieve(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Entity.Retrieve")
	defer span.End()

	accountID, entityID, _ := takeAEI(ctx, params, e.db)
	enty, err := entity.Retrieve(ctx, accountID, entityID, e.db)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, createViewModelEntity(enty), http.StatusOK)
}

// Create inserts a new team into the system.
func (e *Entity) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Entity.Create")
	defer span.End()

	accountID, _, _ := takeAEI(ctx, params, e.db)
	var ne entity.NewEntity
	if err := web.Decode(r, &ne); err != nil {
		return errors.Wrap(err, "")
	}
	ne.ID = uuid.New().String()
	ne.AccountID = accountID
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

	accountID, entityID, _ := takeAEI(ctx, params, e.db)
	var ve entity.ViewModelEntity
	if err := web.Decode(r, &ve); err != nil {
		return errors.Wrap(err, "")
	}

	input, err := json.Marshal(ve.Fields)
	if err != nil {
		return errors.Wrap(err, "encode fields to input")
	}

	err = entity.Update(ctx, e.db, accountID, entityID, string(input), time.Now())
	if err != nil {
		return errors.Wrapf(err, "Entity: %+v", &ve)
	}

	return web.Respond(ctx, w, ve, http.StatusOK)
}

func (e *Entity) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	accountID, entityID, _ := takeAEI(ctx, params, e.db)

	err := entity.Delete(ctx, e.db, accountID, entityID)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, "SUCCESS", http.StatusAccepted)
}

func createViewModelEntity(e entity.Entity) entity.ViewModelEntity {
	return entity.ViewModelEntity{
		ID:            e.ID,
		TeamID:        e.TeamID,
		Name:          e.Name,
		DisplayName:   e.DisplayName,
		Category:      e.Category,
		State:         e.State,
		Fields:        e.DomFields(),
		Tags:          e.Tags,
		IsPublic:      e.IsPublic,
		IsCore:        e.IsCore,
		IsShared:      e.IsShared,
		SharedTeamIds: e.SharedTeamIds,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
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
		log.Println("***> unexpected error occurred. when parsing category_id from the request", err)
		return ids
	} else if i == -1 { // fetch all categories
		return ids
	}

	ids = append(ids, i)
	return ids
}

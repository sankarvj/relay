package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/account"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/reference"
	"gitlab.com/vjsideprojects/relay/internal/team"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
)

// Entity represents the Entity API method handler set.
type Entity struct {
	db            *sqlx.DB
	sdb           *database.SecDB
	authenticator *auth.Authenticator
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

func (e *Entity) Home(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Entity.Home")
	defer span.End()

	role, _ := ctx.Value(auth.RoleKey).(string)
	// if ok {
	// 	log.Println("role--- ", role)
	// }

	acc, err := account.Retrieve(ctx, e.db, params["account_id"])
	if err != nil {
		return err
	}

	cu, err := user.RetrieveCurrentUser(ctx, acc.ID, e.db)
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
		Role           string                    `json:"role"`
		Plan           int                       `json:"plan"`
		Status         string                    `json:"status"`
		TrailEnd       float64                   `json:"trail_end"`
	}{
		acc.Name,
		teamID,
		teams,
		viewModelEntities,
		createViewModelUser(*cu, acc.ID),
		createViewModelUS(cus),
		role,
		acc.CustomerPlan,
		acc.CustomerStatus,
		-math.Trunc(time.Since(util.ConvertMilliToTime(int64(acc.TrailEnd*1000))).Hours() / 24),
	}

	return web.Respond(ctx, w, homeDetail, http.StatusOK)
}

func (e *Entity) Dash(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Entity.Dash")
	defer span.End()

	entities, err := entity.TeamEntities(ctx, params["account_id"], params["team_id"], []int{}, e.db)
	if err != nil {
		return err
	}

	taskEntities := make([]entity.ViewModelEntity, 0)
	approvalEntities := make([]entity.ViewModelEntity, 0)
	streamEntities := make([]entity.ViewModelEntity, 0)
	for _, entt := range entities {
		if entt.Category == entity.CategoryTask {
			taskEntities = append(taskEntities, createViewModelEntity(entt))
		} else if entt.Category == entity.CategoryApprovals {
			approvalEntities = append(approvalEntities, createViewModelEntity(entt))
		} else if entt.Category == entity.CategoryEvent {
			streamEntities = append(streamEntities, createViewModelEntity(entt))
		}
	}

	dashDetail := struct {
		TaskEntities     []entity.ViewModelEntity `json:"task_entities"`
		ApprovalEntities []entity.ViewModelEntity `json:"approval_entities"`
		StreamEntities   []entity.ViewModelEntity `json:"stream_entities"`
	}{
		taskEntities,
		approvalEntities,
		streamEntities,
	}

	return web.Respond(ctx, w, dashDetail, http.StatusOK)
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
	enty, err := entity.Retrieve(ctx, accountID, entityID, e.db, e.sdb)
	if err != nil {
		return err
	}

	fields := enty.AllFieldsButSecured()
	reference.UpdateReferenceFields(ctx, accountID, entityID, fields, []item.Item{}, map[string]interface{}{}, e.db, e.sdb, job.NewJabEngine())

	return web.Respond(ctx, w, createViewModelEntityWithChoices(enty, fields), http.StatusOK)
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
	if ne.Category == entity.CategoryEvent {
		contactEntity, err := entity.RetrieveFixedEntity(ctx, e.db, ne.AccountID, ne.TeamID, entity.FixedEntityContacts)
		if err != nil {
			return err
		}
		ne.Fields = eventFields(contactEntity.ID, contactEntity.Key("first_name"), contactEntity.Key("email"))
	} else if ne.Category == entity.CategoryChildUnit {
		ne.Fields = statusFields()
	} else {
		//add key with a UUID
		fieldKeyBinder(ne.ID, ne.Fields)
	}

	enty, err := entity.Create(ctx, e.db, ne, time.Now())
	if err != nil {
		return errors.Wrapf(err, "Entity: %+v", &enty)
	}

	//cache
	e.sdb.SetEntity(enty.ID, enty.Encode())

	return web.Respond(ctx, w, createViewModelEntity(enty), http.StatusCreated)
}

// Update updates the entity
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

	err = entity.Update(ctx, e.db, e.sdb, accountID, entityID, string(input), time.Now())
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

func createViewModelEntityWithChoices(e entity.Entity, fields []entity.Field) entity.ViewModelEntity {
	return entity.ViewModelEntity{
		ID:            e.ID,
		TeamID:        e.TeamID,
		Name:          e.Name,
		DisplayName:   e.DisplayName,
		Category:      e.Category,
		State:         e.State,
		Fields:        fields,
		Tags:          e.Tags,
		IsPublic:      e.IsPublic,
		IsCore:        e.IsCore,
		IsShared:      e.IsShared,
		SharedTeamIds: e.SharedTeamIds,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
		HasFlow:       e.FlowField() != nil,
		LayoutStyle:   e.Layout(),
	}
}

func createViewModelEntity(e entity.Entity) entity.ViewModelEntity {
	return entity.ViewModelEntity{
		ID:            e.ID,
		TeamID:        e.TeamID,
		Name:          e.Name,
		DisplayName:   e.DisplayName,
		Category:      e.Category,
		State:         e.State,
		Fields:        e.AllFieldsButSecured(),
		Tags:          e.Tags,
		IsPublic:      e.IsPublic,
		IsCore:        e.IsCore,
		IsShared:      e.IsShared,
		SharedTeamIds: e.SharedTeamIds,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
		HasFlow:       e.FlowField() != nil,
		LayoutStyle:   e.Layout(),
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

func (e *Entity) UpdateLS(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Entity.UpdateLS")
	defer span.End()

	ls := params["ls"]
	accountID, entityID, _ := takeAEI(ctx, params, e.db)
	enty, err := entity.Retrieve(ctx, accountID, entityID, e.db, e.sdb)
	if err != nil {
		return err
	}
	existingMeta := enty.Meta()
	existingMeta[entity.MetaRender] = ls
	input, err := json.Marshal(existingMeta)
	if err != nil {
		return errors.Wrap(err, "encode meta to input")
	}
	metab := string(input)
	enty.Metab = &metab
	if ls == entity.MetaRenderPipe {
		err := enty.UpdateMeta(ctx, e.db)
		if err != nil {
			return err
		}
		return web.Respond(ctx, w, createViewModelEntity(enty), http.StatusOK)
	} else if ls == entity.MetaRenderList {
		err := enty.UpdateMeta(ctx, e.db)
		if err != nil {
			return err
		}
		return web.Respond(ctx, w, createViewModelEntity(enty), http.StatusOK)
	} else {
		return web.Respond(ctx, w, "failure", http.StatusBadRequest)
	}
}

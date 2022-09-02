package handlers

import (
	"context"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/stream"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/team"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
)

type Member struct {
	db            *sqlx.DB
	authenticator *auth.Authenticator
	rPool         *redis.Pool
}

func (m *Member) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Member.List")
	defer span.End()

	accountID, entityID, _ := takeAEI(ctx, params, m.db)

	e, err := entity.Retrieve(ctx, accountID, entityID, m.db)
	if err != nil {
		return err
	}

	items, err := item.ListFilterByState(ctx, accountID, entityID, item.StateDefault, m.db)
	if err != nil {
		return err
	}

	teams, err := team.Map(ctx, params["account_id"], m.db)
	if err != nil {
		return err
	}

	viewModelMembers := memberResponse(e, items, teams)

	response := struct {
		EntityID string               `json:"entity_id"`
		Members  []ViewModelMember    `json:"members"`
		TeamMap  map[string]team.Team `json:"team_map"`
	}{
		EntityID: entityID,
		Members:  viewModelMembers,
		TeamMap:  teams,
	}

	return web.Respond(ctx, w, response, http.StatusOK)
}

func (m *Member) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {

	ctx, span := trace.StartSpan(ctx, "handlers.Member.Create")
	defer span.End()

	accountID, entityID, _ := takeAEI(ctx, params, m.db)
	var vm ViewModelMember
	if err := web.Decode(r, &vm); err != nil {
		return errors.Wrap(err, "")
	}

	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return err
	}

	//current entity
	e, err := entity.Retrieve(ctx, accountID, entityID, m.db)
	if err != nil {
		return err
	}
	namedKeys := entity.NamedKeysMap(e.FieldsIgnoreError())

	ni := item.NewItem{
		ID:        uuid.New().String(),
		AccountID: accountID,
		EntityID:  entityID,
		UserID:    &currentUserID,
		Fields:    recreateFields(vm, namedKeys),
	}

	errorMap := validateItemCreate(ctx, accountID, entityID, ni.Fields, m.db, m.rPool)
	if errorMap != nil {
		return web.Respond(ctx, w, errorMap, http.StatusForbidden)
	}

	it, err := createAndPublish(ctx, currentUserID, ni, m.db, m.rPool, m.authenticator.FireBaseAdminSDK)
	if err != nil {
		return errors.Wrapf(err, "Item: %+v", &ni)
	}

	teams, err := team.Map(ctx, params["account_id"], m.db)
	if err != nil {
		return err
	}

	newMember := createViewModelMember(it.ID, entity.NamedFieldsObjMap(e.ValueAdd(it.Fields())), teams)
	return web.Respond(ctx, w, newMember, http.StatusCreated)

}

func (m *Member) Update(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Member.Update")
	defer span.End()

	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return err
	}

	accountID, entityID, memberID := params["account_id"], params["entity_id"], params["member_id"]
	var vm ViewModelMember
	if err := web.Decode(r, &vm); err != nil {
		return errors.Wrap(err, "")
	}
	existingItem, err := item.Retrieve(ctx, entityID, memberID, m.db)
	if err != nil {
		return errors.Wrapf(err, "Member Get During Update")
	}

	//current entity
	e, err := entity.Retrieve(ctx, accountID, entityID, m.db)
	if err != nil {
		return err
	}
	namedKeys := entity.NamedKeysMap(e.FieldsIgnoreError())

	updatedFields := recreateFields(vm, namedKeys)
	it, err := item.UpdateFields(ctx, m.db, entityID, memberID, updatedFields)
	if err != nil {
		return errors.Wrapf(err, "member update: %+v", &it)
	}
	//stream
	go job.NewJob(m.db, m.rPool, m.authenticator.FireBaseAdminSDK).Stream(stream.NewUpdateItemMessage(ctx, m.db, accountID, currentUserID, entityID, memberID, it.Fields(), existingItem.Fields()))

	return web.Respond(ctx, w, createViewModelItem(it), http.StatusOK)
}

func (m *Member) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	accountID, entityID, memberID := params["account_id"], params["entity_id"], params["member_id"]

	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return errors.Wrap(err, "problem reteriving current user")
	}

	ownerEntity, err := entity.RetrieveFixedEntity(ctx, m.db, accountID, "", entity.FixedEntityOwner)
	if err != nil {
		return errors.Wrap(err, "problem reteriving owner entity")
	}

	userItem, err := entity.RetriveUserItem(ctx, accountID, ownerEntity.ID, memberID, m.db)
	if err != nil {
		return errors.Wrap(err, "member id does not exist")
	}

	if currentUserID == userItem.UserID {
		return errors.Wrap(err, "Operation not permitted. Please transfer the ownership of this account or delete the account itself")
	}

	removableUser, err := user.RetrieveUser(ctx, m.db, userItem.UserID)
	if err != nil {
		return err
	}

	err = removableUser.UpdateAccounts(ctx, m.db, removableUser.RemoveAccount(accountID))
	if err != nil {
		return errors.Wrap(err, "removing accounts from the user failed")
	}

	//stream
	go job.NewJob(m.db, m.rPool, m.authenticator.FireBaseAdminSDK).Stream(stream.NewDeleteItemMessage(ctx, m.db, accountID, currentUserID, entityID, memberID))

	return web.Respond(ctx, w, "SUCCESS", http.StatusAccepted)
}

func memberResponse(e entity.Entity, items []item.Item, teamMap map[string]team.Team) []ViewModelMember {
	viewModelItems := make([]ViewModelMember, len(items))
	for i, item := range items {
		viewModelItems[i] = createViewModelMember(item.ID, entity.NamedFieldsObjMap(e.ValueAdd(item.Fields())), teamMap)
	}
	return viewModelItems
}

func createViewModelMember(id string, fields map[string]entity.Field, teamMap map[string]team.Team) ViewModelMember {
	return ViewModelMember{
		ID:     id,
		Name:   fields["name"].Value.(string),
		Email:  fields["email"].Value.(string),
		Avatar: fields["avatar"].Value.(string),
		Teams:  populateTeams(fields["team_ids"].Value, teamMap),
		Role:   fields["role"].Value.([]interface{}),
	}
}

func recreateFields(vm ViewModelMember, namedKeys map[string]string) map[string]interface{} {
	itemFields := make(map[string]interface{}, 0)
	itemFields[namedKeys["name"]] = vm.Name
	itemFields[namedKeys["user_id"]] = vm.UserID
	itemFields[namedKeys["email"]] = vm.Email
	itemFields[namedKeys["avatar"]] = vm.Avatar
	itemFields[namedKeys["team_ids"]] = stripeTeamIds(vm.Teams)
	itemFields[namedKeys["role"]] = vm.Role
	return itemFields
}

type ViewModelMember struct {
	ID     string        `json:"id"`
	UserID string        `json:"user_id"`
	Name   string        `json:"name"`
	Email  string        `json:"email"`
	Avatar string        `json:"avatar"`
	Teams  []ViewTeam    `json:"teams"`
	Role   []interface{} `json:"role"`
}

type ViewTeam struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Cypher struct {
	View   bool `json:"view"`
	Edit   bool `json:"edit"`
	Create bool `json:"create"`
}

func populateTeams(teamIds interface{}, teamMap map[string]team.Team) []ViewTeam {
	vteams := make([]ViewTeam, 0)
	if teamIds == nil {
		return vteams
	}
	for _, tid := range teamIds.([]interface{}) {
		if t, ok := teamMap[tid.(string)]; ok {
			vteams = append(vteams, ViewTeam{ID: t.ID, Name: t.Name})
		}
	}
	return vteams
}

func stripeTeamIds(teams []ViewTeam) []interface{} {
	teamList := make([]interface{}, 0)
	if teams == nil {
		return teamList
	}
	for _, t := range teams {
		teamList = append(teamList, t.ID)
	}
	return teamList
}

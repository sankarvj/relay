package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
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

	items, err := item.ListFilterByState(ctx, entityID, item.StateDefault, m.db)
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

	it, err := createAndPublish(ctx, ni, m.db, m.rPool)
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
	log.Printf("updatedFields %+v", updatedFields)
	it, err := item.UpdateFields(ctx, m.db, entityID, memberID, updatedFields)
	if err != nil {
		log.Println("updatedFields err", err)
		return errors.Wrapf(err, "Item Update: %+v", &it)
	}
	//TODO push this to stream/queue
	(&job.Job{}).EventItemUpdated(accountID, entityID, memberID, it.Fields(), existingItem.Fields(), m.db, m.rPool)

	return web.Respond(ctx, w, createViewModelItem(it), http.StatusOK)
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
		ID:        id,
		Name:      fields["name"].Value.(string),
		Email:     fields["email"].Value.(string),
		Avatar:    fields["avatar"].Value.(string),
		Teams:     populateTeams(fields["team_ids"].Value, teamMap),
		AccessMap: decodeAccessCypher(fields["access_map"].Value),
	}
}

func recreateFields(vm ViewModelMember, namedKeys map[string]string) map[string]interface{} {
	itemFields := make(map[string]interface{}, 0)
	itemFields[namedKeys["name"]] = vm.Name
	itemFields[namedKeys["email"]] = vm.Email
	itemFields[namedKeys["avatar"]] = vm.Avatar
	itemFields[namedKeys["team_ids"]] = stripeTeamIds(vm.Teams)
	itemFields[namedKeys["access_map"]] = marshalAccessMap(vm.AccessMap)
	return itemFields
}

type ViewModelMember struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Email     string            `json:"email"`
	Avatar    string            `json:"avatar"`
	Teams     []ViewTeam        `json:"teams"`
	AccessMap map[string]Cypher `json:"access_map"`
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

func decodeAccessCypher(accessCypherStr interface{}) map[string]Cypher {
	accessCypher := make(map[string]Cypher, 0)
	log.Println("accessCypherStr ", accessCypherStr)
	if accessCypherStr == nil || accessCypherStr == "" || accessCypherStr == "{}" {
		accessCypher["A"] = Cypher{
			View:   true,
			Edit:   true,
			Create: true,
		}
		accessCypher["W"] = Cypher{
			View:   true,
			Edit:   true,
			Create: true,
		}
		log.Println("accessCypher ", accessCypher)
		return accessCypher
	}
	accessMap := unmarshalAccessStr(accessCypherStr.(string))
	// for module, cypherStr := range accessMap {
	// 	c := Cypher{
	// 		View:   false,
	// 		Edit:   false,
	// 		Create: false,
	// 	}
	// 	parts := strings.Split(cypherStr, "_")
	// 	for _, p := range parts {
	// 		switch p {
	// 		case "V":
	// 			c.View = true
	// 		case "E":
	// 			c.Edit = true
	// 		case "C":
	// 			c.Create = true
	// 		}
	// 	}
	// 	accessCypher[module] = c
	// }
	return accessMap
}

func unmarshalAccessStr(accessMapStr string) map[string]Cypher {
	var accessMap map[string]Cypher
	if err := json.Unmarshal([]byte(accessMapStr), &accessMap); err != nil {
		log.Println("***> unexpected error when unmarshaling the access map ", err)
		return make(map[string]Cypher, 0)
	}
	return accessMap
}

func marshalAccessMap(cypherMap map[string]Cypher) string {
	fieldsBytes, err := json.Marshal(cypherMap)
	if err != nil {
		return ""
	}
	return string(fieldsBytes)
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

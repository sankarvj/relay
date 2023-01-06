package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	rg "github.com/redislabs/redisgraph-go"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"gitlab.com/vjsideprojects/relay/internal/team"
	"gitlab.com/vjsideprojects/relay/internal/user"
)

type Piper struct {
	NodeKey        string                     `json:"node_key"`
	Nodes          []node.ViewModelNode       `json:"nodes"`
	Items          map[string][]ViewModelItem `json:"items"`
	Tokens         map[string]string          `json:"tokens"`
	Exps           map[string]string          `json:"exps"`
	CountMap       map[string]map[string]int  `json:"count_map"`
	sourceEntityID string
	sourceItemID   string
}

func nodeStages(ctx context.Context, accountID, flowID string, db *sqlx.DB) ([]node.ViewModelNode, error) {
	nodes, err := node.NodeActorsList(ctx, accountID, flowID, db)
	if err != nil {
		return nil, err
	}

	viewModelNodes := make([]node.ViewModelNode, 0)
	for _, n := range nodes {

		if n.Type == node.Stage {
			viewModelNodes = append(viewModelNodes, createViewModelNodeActor(n))
		}
	}

	return sortNodes(viewModelNodes), nil
}

func itemsResp(ctx context.Context, db *sqlx.DB, accountID string, result *rg.QueryResult) ([]item.Item, error) {
	itemIDs := util.ParseGraphResult(result)
	items, err := item.BulkRetrieveItems(ctx, accountID, itemIDs, db)
	if err != nil {
		return []item.Item{}, err
	}

	return sort(items, itemIDs), nil
}

func sort(items []item.Item, itemIds []interface{}) []item.Item {
	itemMap := make(map[string]item.Item, len(items))
	for i := 0; i < len(items); i++ {
		itemMap[items[i].ID] = items[i]
	}
	sortedItems := make([]item.Item, 0)
	for _, id := range itemIds {
		sortedItems = append(sortedItems, itemMap[id.(string)])
	}
	return sortedItems
}

func selectedTeam(ctx context.Context, accountID, teamID string, db *sqlx.DB) ([]team.Team, string, error) {
	teams, err := team.List(ctx, accountID, db)
	if err != nil {
		return nil, "", err
	}

	var oldTeamID string
	var seletedTeamID string
	for _, t := range teams {
		if t.ID == teamID {
			oldTeamID = t.ID
		}
		if t.ID != t.AccountID {
			seletedTeamID = t.ID
		}
	}
	if oldTeamID != "" {
		seletedTeamID = oldTeamID
	}
	return teams, seletedTeamID, nil
}

func addInnerCondition(approvalEntityID, approvalStatusEntityID, approvalStatusKey string, approalStatusVal interface{}) graphdb.Field {
	return graphdb.Field{
		Value:     []interface{}{""},
		DataType:  graphdb.TypeReference,
		RefID:     approvalEntityID,
		IsReverse: false,
		Field: &graphdb.Field{
			Key:       approvalStatusKey,
			RefID:     approvalStatusEntityID,
			DataType:  graphdb.TypeReference,
			Value:     []interface{}{""},
			IsReverse: false,
			Field: &graphdb.Field{
				Expression: graphdb.Operator("eq"),
				Key:        "id",
				DataType:   graphdb.TypeString,
				Value:      approalStatusVal,
			},
		},
	}
}

func makeConditionsFromExp(ctx context.Context, accountID, entityID, exp string, db *sqlx.DB, sdb *database.SecDB) ([]graphdb.Field, error) {
	conditionFields := make([]graphdb.Field, 0)

	filter := job.NewJabEngine().RunExpGrapher(ctx, db, sdb, accountID, exp)

	if filter != nil {
		e, err := entity.Retrieve(ctx, accountID, entityID, db, sdb)
		if err != nil {
			return nil, err
		}

		fields := e.OnlyVisibleFields()

		fieldsMap := entity.KeyMap(fields)

		for k, condition := range filter.Conditions {
			if f, ok := fieldsMap[k]; ok && e.ID == condition.EntityID {
				conditionFields = append(conditionFields, f.MakeGraphField(condition.Term, condition.Expression, false))
			} else {
				eIn, err := entity.Retrieve(ctx, accountID, condition.EntityID, db, sdb)
				if err != nil {
					return nil, err
				}

				fieldsMapIn := entity.KeyMap(eIn.EasyFields())
				if f, ok := fieldsMapIn[k]; ok {
					cn := addInnerCondition(condition.EntityID, f.RefID, k, condition.Term)
					conditionFields = append(conditionFields, cn)
				}

			}
		}
	}
	return conditionFields, nil
}

func validateMyself(ctx context.Context, accountID, entityID, itemID string, db *sqlx.DB) error {
	validateMyself, _ := ctx.Value(auth.ValidateMyItemKey).(bool)
	if validateMyself {
		currentUserID, err := user.RetrieveCurrentUserID(ctx)
		if err != nil {
			return err
		}
		it, err := item.Retrieve(ctx, accountID, entityID, itemID, db)
		if err != nil {
			return err
		}
		if it.UserID == nil || *it.UserID != currentUserID {
			err := errors.New("you_dont_have_access_todo_this_operation")
			return web.NewRequestError(err, http.StatusForbidden)
		}
	}
	return nil
}

func validateMyselfWithOwner(ctx context.Context, accountID, entityID string, item item.Item, db *sqlx.DB, sdb *database.SecDB) error {
	validateMyself, _ := ctx.Value(auth.ValidateMyItemKey).(bool)
	if validateMyself {
		currentUserID, err := user.RetrieveCurrentUserID(ctx)
		if err != nil {
			return err
		}

		if item.UserID == nil || *item.UserID != currentUserID {
			err := errors.New("you_dont_have_access_todo_this_operation")
			return web.NewRequestError(err, http.StatusForbidden)
		}

		cuser, err := user.RetrieveUser(ctx, db, accountID, currentUserID)
		if err != nil {
			err := errors.New("you_dont_have_access_todo_this_operation")
			return web.NewRequestError(err, http.StatusForbidden)
		}

		e, err := entity.Retrieve(ctx, accountID, entityID, db, sdb)
		if err != nil {
			return err
		}

		valueAddedFields := e.ValueAdd(item.Fields())

		var isUserAssigned bool
		for _, vf := range valueAddedFields {
			if vf.Who == entity.WhoAssignee {
				vals := vf.Value.([]interface{})
				for _, v := range vals {
					if v == cuser.MemberID {
						isUserAssigned = true
					}
				}
			}
		}

		if !isUserAssigned {
			err := errors.New("you_dont_have_access_todo_this_operation")
			return web.NewRequestError(err, http.StatusForbidden)
		}
	}
	return nil
}

func eventFields(contactEntityID, contactEntityKey, contactEntityEmailKey string) []entity.Field {
	activityNameFieldID := uuid.New().String()
	activityNameField := entity.Field{
		Key:         activityNameFieldID,
		Name:        "name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
		Who:         entity.WhoTitle,
	}

	activityDescFieldID := uuid.New().String()
	activityDescField := entity.Field{
		Key:         activityDescFieldID,
		Name:        "description",
		DisplayName: "Description",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutSubTitle},
		Who:         entity.WhoDesc,
	}

	timeOfEventFieldID := uuid.New().String()
	timeOfEventField := entity.Field{
		Key:         timeOfEventFieldID,
		Name:        "time",
		DisplayName: "Time",
		DomType:     entity.DomText,
		DataType:    entity.TypeDateTime,
		Who:         entity.WhoStartTime,
	}

	iconFieldID := uuid.New().String()
	iconField := entity.Field{
		Key:         iconFieldID,
		Name:        "icon",
		DisplayName: "Icon",
		DomType:     entity.DomImage,
		DataType:    entity.TypeString,
		Who:         entity.WhoIcon,
	}

	contactsFieldID := uuid.New().String()
	contactsField := entity.Field{
		Key:         contactsFieldID,
		Name:        "associated_contacts",
		DisplayName: "Associated Contacts",
		DomType:     entity.DomAutoComplete,
		DataType:    entity.TypeReference,
		RefID:       contactEntityID,
		Meta:        map[string]string{entity.MetaKeyDisplayGex: contactEntityKey, entity.MetaKeyEmailGex: contactEntityEmailKey, entity.MetaMultiChoice: "true"},
		Field: &entity.Field{
			DataType: entity.TypeString,
			Key:      "id",
			Value:    "--",
		},
		Who: entity.WhoContacts,
	}

	return []entity.Field{activityNameField, activityDescField, timeOfEventField, contactsField, iconField}
}

func statusFields() []entity.Field {
	verbFieldID := uuid.New().String()
	verbField := entity.Field{
		Key:         verbFieldID, // we use this value inside the code. don't change it
		Name:        "verb",
		DisplayName: "Verb (Internal field)",
		DomType:     entity.DomNotApplicable,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutVerb},
		Who:         entity.WhoVerb,
	}

	nameFieldID := uuid.New().String()
	nameField := entity.Field{
		Key:         nameFieldID,
		Name:        "name",
		DisplayName: "Name",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutTitle},
	}

	colorFieldID := uuid.New().String()
	colorField := entity.Field{
		Key:         colorFieldID,
		Name:        "color",
		DisplayName: "Color",
		DomType:     entity.DomText,
		DataType:    entity.TypeString,
		Meta:        map[string]string{entity.MetaKeyLayout: entity.MetaLayoutColor},
		Who:         entity.WhoColor,
	}

	return []entity.Field{verbField, nameField, colorField}
}

//b.accountID
// func ItemAdd(ctx context.Context, accountID, entityID, itemID, userID string, fields map[string]interface{}, source map[string][]string, db *sqlx.DB, sdb *database.SecDB, firebaseSDKPath string) (item.Item, error) {
// 	name := "System Generated"
// 	ni := item.NewItem{
// 		ID:        itemID,
// 		Name:      &name,
// 		AccountID: accountID,
// 		EntityID:  entityID,
// 		UserID:    &userID,
// 		Fields:    fields,
// 		Source:    source,
// 		Type:      item.TypeDummy,
// 	}

// 	it, err := item.Create(ctx, db, ni, time.Now())
// 	if err != nil {
// 		return item.Item{}, err
// 	}

// 	job.NewJob(db, sdb, firebaseSDKPath).Stream(stream.NewCreteItemMessage(ctx, db, accountID, userID, entityID, it.ID, ni.Source))

// 	return it, nil
// }

func sortNodes(nodes []node.ViewModelNode) []node.ViewModelNode {
	//make map of nodes by its parent_id
	mapOfParentNodes := make(map[string]node.ViewModelNode, 0)
	for _, n := range nodes {
		mapOfParentNodes[n.ParentNodeID] = n
	}

	//pick first node
	firstNode := mapOfParentNodes["00000000-0000-0000-0000-000000000000"]
	if firstNode.ID == "" {
		return nodes // atleast send the nodes when caught with error
	}

	//pick first node
	sortedNodes := make([]node.ViewModelNode, len(nodes))
	sortedNodes[0] = firstNode
	parentNodeID := firstNode.ID

	for i := 1; i < len(nodes); i++ {
		if nextNode, ok := mapOfParentNodes[parentNodeID]; ok {
			sortedNodes[i] = nextNode
			parentNodeID = nextNode.ID
		}
	}

	return sortedNodes
}

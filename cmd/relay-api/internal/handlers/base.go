package handlers

import (
	"context"
	"log"
	"net/http"

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
	return viewModelNodes, nil
}

func itemsResp(ctx context.Context, db *sqlx.DB, accountID string, result *rg.QueryResult) ([]item.Item, error) {
	itemIDs := util.ParseGraphResult(result)
	items, err := item.BulkRetrieveItems(ctx, accountID, itemIDs, db)
	if err != nil {
		return []item.Item{}, err
	}

	return sort(items, itemIDs), nil
}

func itemElements(result *rg.QueryResult) []interface{} {
	values := make([]interface{}, 0)
	for result.Next() { // Next returns true until the iterator is depleted.
		// Get the current Record.
		r := result.Record()

		// Entries in the Record can be accessed by index or key.
		record := util.ConvertInterfaceToMap(util.ConvertInterfaceToMap(r.GetByIndex(0))["Properties"])
		log.Printf("record %+v", record)
		values = append(values, record["element"])
	}
	return values
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

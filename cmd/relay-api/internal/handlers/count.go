package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	rg "github.com/redislabs/redisgraph-go"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
)

type Counter struct {
	db    *sqlx.DB
	rPool *redis.Pool
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// started with the not so generic manner. Will add the generic later
func (c *Counter) Count(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	var countBody CountRequest
	if err := web.Decode(r, &countBody); err != nil {
		return errors.Wrap(err, "")
	}
	accountID := params["account_id"]
	teamID := params["team_id"]
	entityID := params["entity_id"]      //deal entity
	destination := params["destination"] //entity.FixedEntityTask --> tasks/nodes

	dstEntity, err := entity.RetrieveFixedEntity(ctx, c.db, accountID, teamID, destination)
	if err != nil {
		return err
	}

	switch destination {
	case entity.FixedEntityTask:
		res, err := taskCountPerItem(ctx, accountID, entityID, dstEntity, countBody, c.db, c.rPool)
		if err != nil {
			return err
		}
		return web.Respond(ctx, w, res, http.StatusOK)
	case entity.FixedEntityNode:
		res, err := itemCountPerStage(accountID, entityID, dstEntity, countBody, c.rPool)
		if err != nil {
			return err
		}
		return web.Respond(ctx, w, res, http.StatusOK)
	}
	return web.Respond(ctx, w, fmt.Sprintf("%s Not Implemented", destination), http.StatusNotImplemented)
}

func taskCountPerItem(ctx context.Context, accountID, entityID string, dstEntity entity.Entity, countBody CountRequest, db *sqlx.DB, rPool *redis.Pool) (map[string]CountResponse, error) {
	var statusField entity.Field
	for _, f := range dstEntity.FieldsIgnoreError() {
		if f.Who == entity.WhoStatus {
			statusField = f
			break
		}
	}

	conditionFieldsForAll := makeConditionFieldForAll(countBody.IDs, dstEntity, statusField)
	doneID, _ := entity.DiscoverDoneStatusID(ctx, accountID, statusField.RefID, db)
	conditionFieldsForDone := makeConditionFieldForDone(countBody.IDs, doneID, dstEntity, statusField)

	gSegmentA := graphdb.BuildGNode(accountID, entityID, false).MakeBaseGNode("", conditionFieldsForAll)
	resultA, err := graphdb.GetCount(rPool, gSegmentA, true, true)
	if err != nil {
		return nil, err
	}
	allTasksCount := counts(resultA)

	gSegmentD := graphdb.BuildGNode(accountID, entityID, false).MakeBaseGNode("", conditionFieldsForDone)
	resultD, err := graphdb.GetCount(rPool, gSegmentD, true, true)
	if err != nil {
		return nil, err
	}
	doneTasksCount := counts(resultD)
	return countsResponse(allTasksCount, doneTasksCount), nil
}

//itemCountPerStage
func itemCountPerStage(accountID, entityID string, dstEntity entity.Entity, countBody CountRequest, rPool *redis.Pool) (map[string]int, error) {
	conditionFieldsForStage := makeItemPerStage(dstEntity, countBody.IDs)

	gSegmentA := graphdb.BuildGNode(accountID, entityID, false).MakeBaseGNode("", conditionFieldsForStage)
	resultA, err := graphdb.GetCount(rPool, gSegmentA, false, true)
	if err != nil {
		return nil, err
	}
	return counts(resultA), nil
}

func counts(result *rg.QueryResult) map[string]int {
	responseArr := make(map[string]int, 0)
	for result.Next() { // Next returns true until the iterator is depleted.
		// Get the current Record.

		r := result.Record()

		id := "total_count"
		if len(r.Keys()) > 1 {
			id = r.GetByIndex(1).(string)
		}
		count := r.GetByIndex(0).(int)
		responseArr[id] = count
	}
	return responseArr
}

func countsResponse(allMap, doneMap map[string]int) map[string]CountResponse {
	responseArr := make(map[string]CountResponse, 0)

	for key, acount := range allMap {
		if dCount, ok := doneMap[key]; ok {
			responseArr[key] = CountResponse{All: acount, Done: dCount}
		} else {
			responseArr[key] = CountResponse{All: acount, Done: 0}
		}
	}
	return responseArr
}

func makeConditionFieldForAll(ids []string, dstEntity entity.Entity, statusField entity.Field) []graphdb.Field {
	conditionFields := []graphdb.Field{
		{
			Expression: "in", //adding IN instead of giving the ID in the MakeBaseGNode
			Key:        "id",
			DataType:   graphdb.TypeWist,
			Value:      ids,
		},
		{
			Value:    []interface{}{""}, //this makes the relation between src and dst entity
			RefID:    dstEntity.ID,
			DataType: graphdb.TypeReference,
			Field:    &graphdb.Field{ // this adds the condition to the relation over the task
			},
		},
	}
	return conditionFields
}

func makeConditionFieldForDone(ids []string, doneID string, dstEntity entity.Entity, statusField entity.Field) []graphdb.Field {
	conditionFields := []graphdb.Field{
		{
			Expression: "in", //adding IN instead of giving the ID in the MakeBaseGNode
			Key:        "id",
			DataType:   graphdb.TypeWist,
			Value:      ids,
		},
		{
			Value:    []interface{}{""}, //this makes the relation between src and dst entity
			RefID:    dstEntity.ID,
			DataType: graphdb.TypeReference,
			Field: &graphdb.Field{ // this adds the condition to the relation over the task
				RefID:    statusField.RefID,
				DataType: graphdb.TypeReference,
				Value:    []interface{}{""},
				Field: &graphdb.Field{
					Expression: "=",
					Key:        "id",
					DataType:   graphdb.TypeString,
					Value:      doneID, // status verb as done
				},
			},
		},
	}
	return conditionFields
}

func makeItemPerStage(dstEntity entity.Entity, ids []string) []graphdb.Field {
	conditionFields := []graphdb.Field{
		{
			Value:    []interface{}{""}, //this makes the relation between src and dst entity
			RefID:    dstEntity.ID,
			DataType: graphdb.TypeReference,
			Field: &graphdb.Field{ // this adds the condition to the relation over the dst entity
				Expression: "in", //adding IN instead of giving the ID in the MakeBaseGNode
				Key:        "id",
				DataType:   graphdb.TypeWist,
				Value:      ids,
			},
		},
	}
	return conditionFields
}

type CountRequest struct {
	IDs []string `json:"ids"`
}

type CountResponse struct {
	Done int `json:"done"`
	All  int `json:"all"`
}

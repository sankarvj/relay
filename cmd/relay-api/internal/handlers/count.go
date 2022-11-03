package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	rg "github.com/redislabs/redisgraph-go"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/reference"
)

type Counter struct {
	db  *sqlx.DB
	sdb *database.SecDB
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
		res, err := taskCountPerItem(ctx, accountID, entityID, dstEntity, countBody, c.db, c.sdb)
		if err != nil {
			return err
		}
		return web.Respond(ctx, w, res, http.StatusOK)
	case entity.FixedEntityNode:
		res, err := recordCountPerStage(accountID, entityID, dstEntity, countBody, c.sdb)
		if err != nil {
			return err
		}
		return web.Respond(ctx, w, res, http.StatusOK)
	}
	return web.Respond(ctx, w, fmt.Sprintf("%s Not Implemented", destination), http.StatusNotImplemented)
}

func taskCountPerItem(ctx context.Context, accountID, entityID string, dstEntity entity.Entity, countBody CountRequest, db *sqlx.DB, sdb *database.SecDB) (map[string][]Series, error) {
	conditionFields := make([]graphdb.Field, 0)
	statusField := dstEntity.WhoField(entity.WhoStatus)

	conditionFields = append(conditionFields, makeIdCondition(countBody.IDs, dstEntity.ID, statusField.RefID)...)

	gSegment := graphdb.BuildGNode(accountID, entityID, false, nil).MakeBaseGNode("", conditionFields)
	result, err := graphdb.GetGroupedIDPlusFieldCount(sdb.GraphPool(), gSegment, statusField.RefID, true)
	if err != nil {
		return nil, err
	}
	reference.LoadRefFieldChoices(ctx, accountID, &statusField, db, sdb)

	response := make(map[string][]Series, 0)
	for result.Next() {
		r := result.Record()
		itemID := r.GetByIndex(1).(string)
		statusID := r.GetByIndex(2).(string)

		choice := statusField.ChoiceMap()[statusID]
		if _, ok := response[itemID]; !ok {
			response[itemID] = make([]Series, 0)
		}

		switch v := r.GetByIndex(0).(type) {
		case int:
			response[itemID] = append(response[itemID], createPartialVMSeries(choice.ID, choice.DisplayValue.(string), choice.Color, choice.Verb, v))
		case float64:
			response[itemID] = append(response[itemID], createPartialVMSeries(choice.ID, choice.DisplayValue.(string), choice.Color, choice.Verb, int(v)))
		}
	}

	return response, nil
}

func recordCountPerStage(accountID, entityID string, dstEntity entity.Entity, countBody CountRequest, sdb *database.SecDB) (map[string]int, error) {
	conditionFieldsForStage := makeItemPerStage(dstEntity, countBody.IDs)

	gSegment := graphdb.BuildGNode(accountID, entityID, false, nil).MakeBaseGNode("", conditionFieldsForStage)
	result, err := graphdb.GetCount(sdb.GraphPool(), gSegment, true)
	if err != nil {
		return nil, err
	}
	return counts(result), nil
}

func (gOne *GridOne) itemCountPerStage(ctx context.Context, accountID, exp string, fields []entity.Field, db *sqlx.DB, sdb *database.SecDB) error {
	conditionFields := make([]graphdb.Field, 0)

	for _, f := range fields {
		if f.IsFlow() {
			conditionFields = append(conditionFields, conditionableRef(f, gOne.SelectedFlowID))
		}
		if f.IsNode() {
			conditionFields = append(conditionFields, *f.MakeGraphFieldPlain())
		}
	}

	filter := job.NewJabEngine().RunExpGrapher(ctx, db, sdb, accountID, exp)
	if filter != nil {
		for _, f := range fields {
			if condition, ok := filter.Conditions[f.Key]; ok {
				conditionFields = append(conditionFields, f.MakeGraphField(condition.Term, condition.Expression, false))
			}
		}
	}

	gSegment := graphdb.BuildGNode(accountID, gOne.SelectedEntityID, false, nil).MakeBaseGNode("", conditionFields)

	result, err := graphdb.GetCount(sdb.GraphPool(), gSegment, true)
	if err != nil {
		return errors.Wrapf(err, "WidgetGridOne: get stage count")
	}
	gOne.gridResult(ctx, counts(result), db)
	return nil
}

func (gOne *GridOne) taskCountPerStage(ctx context.Context, accountID, teamID string, db *sqlx.DB, sdb *database.SecDB) error {
	conditionFields := make([]graphdb.Field, 0)

	taskE, err := entity.RetrieveFixedEntity(ctx, db, accountID, teamID, "tasks")
	if err != nil {
		return errors.Wrapf(err, "WidgetGridThree: entity retieve error")
	}

	var stageFieldKey string
	var statusField entity.Field
	fields := taskE.EasyFields()
	for _, f := range fields {
		if f.Name == "pipeline_stage" {
			stageFieldKey = taskE.Key("pipeline_stage")
		}
		if f.Who == entity.WhoStatus {
			statusField = f
			conditionFields = append(conditionFields, *f.MakeGraphFieldPlain())
		}
	}

	exp := fmt.Sprintf("{{%s.%s}} in {%s}", taskE.ID, stageFieldKey, gOne.stageIds())

	filter := job.NewJabEngine().RunExpGrapher(ctx, db, sdb, accountID, exp)
	if filter != nil {
		for _, f := range fields {
			if condition, ok := filter.Conditions[f.Key]; ok {
				conditionFields = append(conditionFields, f.MakeGraphField(condition.Term, condition.Expression, false))
			}
		}
	}

	gSegment := graphdb.BuildGNode(accountID, taskE.ID, false, nil).MakeBaseGNode("", conditionFields)

	result, err := graphdb.GetGroupedIDPlusFieldCount(sdb.GraphPool(), gSegment, statusField.RefID, false)
	if err != nil {
		return errors.Wrapf(err, "WidgetGridOne: get stage count")
	}
	log.Printf("result resultresultresultresultresultresultresult %+v", result)

	reference.LoadRefFieldChoices(ctx, accountID, &statusField, db, sdb)

	gOne.taskCountForEachStage(ctx, statusField, result, db)
	return nil
}

func counts(result *rg.QueryResult) map[string]int {
	responseArr := make(map[string]int, 0)
	if result == nil {
		return responseArr
	}
	for result.Next() { // Next returns true until the iterator is depleted.
		// Get the current Record.

		r := result.Record()

		id := "total_count"
		if len(r.Keys()) > 1 {
			id = r.GetByIndex(1).(string)
		}

		switch v := r.GetByIndex(0).(type) {
		case int:
			responseArr[id] = v
		case float64:
			responseArr[id] = int(v)
		}
	}

	return responseArr
}

func makeIdCondition(ids []string, dstEntityID, statusEntityID string) []graphdb.Field {
	conditionFields := []graphdb.Field{
		{
			Expression: "in", //adding IN instead of giving the ID in the MakeBaseGNode
			Key:        "id",
			DataType:   graphdb.TypeWist,
			Value:      ids,
		},
		{
			Value:    []interface{}{""}, //this makes the relation between src and dst entity
			RefID:    dstEntityID,
			DataType: graphdb.TypeReference,
			Field: &graphdb.Field{ // this adds the condition to the relation over the task
				RefID:    statusEntityID,
				DataType: graphdb.TypeReference,
				Value:    []interface{}{""},
				Field:    &graphdb.Field{},
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

func (gOne *GridOne) stageIds() string {
	ids := make([]string, len(gOne.Stages))
	for index, stage := range gOne.Stages {
		ids[index] = stage.ID
	}
	return strings.Join(ids[:], ",")
}

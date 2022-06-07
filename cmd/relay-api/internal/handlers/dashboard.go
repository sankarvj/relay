package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/ruler"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/reference"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"go.opencensus.io/trace"
)

type Dashboard struct {
	db    *sqlx.DB
	rPool *redis.Pool
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

type GridOne struct {
	AvailableEntities  []entity.Entity      `json:"available_entities"`
	SelectedEntity     entity.Entity        `json:"selected_entity"`
	SelectedFlow       flow.ViewModelFlow   `json:"selected_flow"`
	SelectedNode       node.ViewModelNode   `json:"selected_node"`
	Flows              []flow.ViewModelFlow `json:"flows"`
	Nodes              []node.ViewModelNode `json:"nodes"`
	SelectedStageCount int                  `json:"selected_stage_count"`
	OtherStageCount    int                  `json:"other_stage_count"`
}

type GridTwo struct {
	AvailableEntities []entity.Entity `json:"available_entities"`
	SelectedEntity    entity.Entity   `json:"selected_entity"`
	Name              string          `json:"name"`
	Count             map[string]int  `json:"count"`
}

type GridThree struct {
	Choices []entity.Choice `json:"choices"`
}

func (d *Dashboard) Overview(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	entities, err := entity.All(ctx, params["account_id"], []int{entity.CategoryData}, d.db)
	if err != nil {
		return err
	}

	availableEntities := make([]entity.Entity, 0)
	for _, e := range entities {
		if e.FlowField() != nil {
			availableEntities = append(availableEntities, e)
		}
	}

	gridOne := GridOne{AvailableEntities: availableEntities}
	err = gridOne.WidgetGridOne(ctx, params["account_id"], params["team_id"], d.db, d.rPool)
	if err != nil {
		return err
	}
	gridTwo := GridTwo{AvailableEntities: availableEntities}
	err = gridTwo.WidgetGridTwo(ctx, params["account_id"], params["team_id"], d.db, d.rPool)
	if err != nil {
		return err
	}
	gridThree := GridThree{}
	err = gridThree.WidgetGridThree(ctx, params["account_id"], params["team_id"], d.db, d.rPool)
	if err != nil {
		return err
	}

	overview := struct {
		GOne   GridOne   `json:"g_one"`
		GTwo   GridTwo   `json:"g_two"`
		GThree GridThree `json:"g_three"`
	}{
		gridOne,
		gridTwo,
		gridThree,
	}
	return web.Respond(ctx, w, overview, http.StatusOK)
}

func (gOne *GridOne) WidgetGridOne(ctx context.Context, accountID, teamID string, db *sqlx.DB, rPool *redis.Pool) error {

	if len(gOne.AvailableEntities) == 0 {
		return nil
	}
	gOne.SelectedEntity = gOne.AvailableEntities[0]
	conditionFields := make([]graphdb.Field, 0)

	if gOne.SelectedEntity.FlowField() != nil { //main stages. ex: deal stages
		flows, err := flow.List(ctx, []string{gOne.SelectedEntity.ID}, flow.FlowModePipeLine, flow.FlowTypeAll, db)
		if err != nil {
			return err
		}

		gOne.Flows = make([]flow.ViewModelFlow, len(flows))
		for i, flow := range flows {
			gOne.Flows[i] = createViewModelFlow(flow, nil)
		}

		if len(flows) > 0 {
			gOne.SelectedFlow = createViewModelFlow(flows[0], nil)
			gOne.Nodes, err = nodeStages(ctx, gOne.SelectedFlow.ID, db)
			if err != nil {
				return err
			}
			if len(gOne.Nodes) > 0 {
				gOne.SelectedNode = gOne.Nodes[len(gOne.Nodes)-1]
			}
		}
	}

	fields, err := gOne.SelectedEntity.FilteredFields()
	if err != nil {
		return errors.Wrapf(err, "WidgetGridOne: fields retieve error")
	}

	for _, f := range fields {
		if f.IsFlow() {
			conditionFields = append(conditionFields, conditionable(f, gOne.SelectedFlow.ID))
		}
		if f.IsNode() {
			conditionFields = append(conditionFields, relatable(f))
		}
	}

	gSegment := graphdb.BuildGNode(accountID, gOne.SelectedEntity.ID, false).MakeBaseGNode("", conditionFields)

	result, err := graphdb.GetCount(rPool, gSegment, true)
	if err != nil {
		return errors.Wrapf(err, "WidgetGridOne: get stage count")
	}
	gOne.gridResult(counts(result))
	log.Printf("stage result %+v	", result)
	return nil
}

func (gTwo *GridTwo) WidgetGridTwo(ctx context.Context, accountID, teamID string, db *sqlx.DB, rPool *redis.Pool) error {
	if len(gTwo.AvailableEntities) == 0 {
		return nil
	}
	gTwo.SelectedEntity = gTwo.AvailableEntities[0]

	conditionFields := make([]graphdb.Field, 0)

	fields, err := gTwo.SelectedEntity.FilteredFields()
	if err != nil {
		return errors.Wrapf(err, "WidgetGridTwo: fields retieve error")
	}

	namedKeyMap := entity.NamedKeysMap(fields)
	var key string
	if gTwo.SelectedEntity.Name == "deals" {
		key = namedKeyMap["deal_amount"]
	} else if gTwo.SelectedEntity.Name == "projects" {
		key = namedKeyMap["uuid-00-status"]
	} else {
		return nil
	}

	//{Operator:in Key:uuid-00-contacts DataType:S Value:6eb4f58e-8327-4ccc-a262-22ad809e76cb}
	gSegment := graphdb.BuildGNode(accountID, gTwo.SelectedEntity.ID, false).MakeBaseGNode("", conditionFields)
	result, err := graphdb.GetSum(rPool, gSegment, key)
	if err != nil {
		return errors.Wrapf(err, "WidgetGridTwo: get amount sum")
	}
	gTwo.gridResult(counts(result))
	log.Printf("result %+v	", result)
	return nil
}

func (gThree *GridThree) WidgetGridThree(ctx context.Context, accountID, teamID string, db *sqlx.DB, rPool *redis.Pool) error {
	conditionFields := make([]graphdb.Field, 0)
	e, err := entity.RetrieveFixedEntity(ctx, db, accountID, teamID, "tasks")
	if err != nil {
		return errors.Wrapf(err, "WidgetGridThree: entity retieve error")
	}

	fields, err := e.FilteredFields()
	if err != nil {
		return errors.Wrapf(err, "WidgetGridThree: fields retieve error")
	}

	var statusField entity.Field
	for _, f := range fields {
		if f.Who == entity.WhoStatus {
			statusField = f
			conditionFields = append(conditionFields, relatable(f))
		}
	}

	//{Operator:in Key:uuid-00-contacts DataType:S Value:6eb4f58e-8327-4ccc-a262-22ad809e76cb}
	gSegment := graphdb.BuildGNode(accountID, e.ID, false).MakeBaseGNode("", conditionFields)

	result, err := graphdb.GetCount(rPool, gSegment, true)
	if err != nil {
		return errors.Wrapf(err, "WidgetGridThree: get status count")
	}
	gThree.gridResult(ctx, accountID, teamID, statusField, counts(result), db)
	log.Println("three result", result)
	return nil
}

func (d *Dashboard) Dashboard(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Dashboard.Overview")
	defer span.End()

	exp := r.URL.Query().Get("exp")

	e, err := entity.RetrieveFixedEntity(ctx, d.db, params["account_id"], params["team_id"], "tasks")
	if err == nil {
		return err
	}

	fields := e.FieldsIgnoreError()
	conditionFields := make([]graphdb.Field, 0)
	whoFieldsMap := entity.WhoFieldsMap(fields)

	filter := overdue(whoFieldsMap, entity.WhoDueBy)

	for _, f := range fields {
		if condition, ok := filter.Conditions[f.Key]; ok {
			conditionFields = append(conditionFields, makeGraphField(&f, condition.Term, condition.Expression, false))
		}
	}

	_, countMap, err := NewSegmenter(exp).
		AddCount().
		filterWrapper(ctx, params["account_id"], e.ID, fields, map[string]interface{}{}, d.db, d.rPool)
	if err != nil {
		return err
	}

	log.Println("countMap---> ", countMap)

	return web.Respond(ctx, w, "", http.StatusOK)
}

func overdue(fields map[string]entity.Field, who string) *ruler.Filter {
	f := fields[who]
	filter := &ruler.Filter{
		Conditions: map[string]ruler.Condition{},
	}
	condition := ruler.Condition{
		Expression: lexertoken.GTSign,
		Term:       util.GetMilliSecondsFloat(time.Now()),
		DataType:   ruler.DType(f.DataType),
	}
	filter.Conditions[f.Key] = condition

	return filter
}

func relatable(f entity.Field) graphdb.Field {
	return graphdb.Field{
		Key:      f.Key,
		Value:    []interface{}{""},
		DataType: graphdb.TypeReference,
		RefID:    f.RefID,
		Field:    &graphdb.Field{},
	}
}

func conditionable(f entity.Field, value interface{}) graphdb.Field {
	return graphdb.Field{
		Key:      f.Key,
		Value:    []interface{}{value},
		DataType: graphdb.TypeReference,
		RefID:    f.RefID,
		Field:    &graphdb.Field{},
	}
}

func (gOne *GridOne) gridResult(resultMap map[string]int) {
	for k, v := range resultMap {
		if k == gOne.SelectedNode.ID {
			gOne.SelectedStageCount = v
		} else {
			gOne.OtherStageCount = gOne.OtherStageCount + v
		}
	}
}

func (gTwo *GridTwo) gridResult(resultMap map[string]int) {
	gTwo.Name = "Total ARR"
	gTwo.Count = resultMap
}

func (gThree *GridThree) gridResult(ctx context.Context, accountID, teamID string, f entity.Field, resultMap map[string]int, db *sqlx.DB) {
	e, _ := entity.Retrieve(ctx, accountID, f.RefID, db)
	refItems, _ := item.EntityItems(ctx, e.ID, db)

	choicer := reference.ItemChoices(&f, refItems, e.WhoFields())
	gThree.Choices = make([]entity.Choice, 0)

	for i := 0; i < len(choicer); i++ {
		if val, ok := resultMap[choicer[i].ID]; ok {
			gThree.Choices = append(gThree.Choices, entity.Choice{
				ID:           choicer[i].ID,
				DisplayValue: choicer[i].Name,
				Value:        val,
				Verb:         util.ConvertIntfToStr(choicer[i].Verb),
				Avatar:       choicer[i].Avatar,
			})
		} else {
			gThree.Choices = append(gThree.Choices, entity.Choice{
				ID:           choicer[i].ID,
				DisplayValue: choicer[i].Name,
				Value:        0,
				Verb:         util.ConvertIntfToStr(choicer[i].Verb),
				Avatar:       choicer[i].Avatar,
			})
		}

	}
}

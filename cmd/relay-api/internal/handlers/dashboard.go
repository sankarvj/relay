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
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/ruler"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
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
	SelectedFlow flow.ViewModelFlow   `json:"selected_flow"`
	SelectedNode node.ViewModelNode   `json:"selected_node"`
	Flows        []flow.ViewModelFlow `json:"flows"`
	Nodes        []node.ViewModelNode `json:"nodes"`
}

func (d *Dashboard) Overview(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	WidgetDeals(ctx, params["account_id"], params["team_id"], d.db, d.rPool)
	return web.Respond(ctx, w, "", http.StatusOK)
}

func WidgetDeals(ctx context.Context, accountID, teamID string, db *sqlx.DB, rPool *redis.Pool) error {
	gridOne := GridOne{}
	conditionFields := make([]graphdb.Field, 0)

	e, err := entity.RetrieveFixedEntity(ctx, db, accountID, teamID, "deals")
	if err != nil {
		return errors.Wrapf(err, "WidgetTasks: entity retieve error")
	}

	if e.FlowField() != nil { //main stages. ex: deal stages
		flows, err := flow.List(ctx, []string{e.ID}, flow.FlowModePipeLine, flow.FlowTypeAll, db)
		if err != nil {
			return err
		}

		gridOne.Flows = make([]flow.ViewModelFlow, len(flows))
		for i, flow := range flows {
			gridOne.Flows[i] = createViewModelFlow(flow, nil)
		}

		if len(flows) > 0 {
			gridOne.SelectedFlow = createViewModelFlow(flows[0], nil)
			gridOne.Nodes, err = nodeStages(ctx, gridOne.SelectedFlow.ID, db)
			if err != nil {
				return err
			}
			if len(gridOne.Nodes) > 0 {
				gridOne.SelectedNode = gridOne.Nodes[len(gridOne.Nodes)-1]
			}
		}

	}

	fields, err := e.FilteredFields()
	if err != nil {
		return errors.Wrapf(err, "WidgetTasks: fields retieve error")
	}

	for _, f := range fields {
		if f.IsFlow() {
			conditionFields = append(conditionFields, conditionable(f, gridOne.SelectedFlow.ID))
		}
		if f.IsNode() {
			conditionFields = append(conditionFields, conditionable(f, gridOne.SelectedNode.ID))
		}
	}

	//{Operator:in Key:uuid-00-contacts DataType:S Value:6eb4f58e-8327-4ccc-a262-22ad809e76cb}
	gSegment := graphdb.BuildGNode(accountID, e.ID, false).MakeBaseGNode("", conditionFields)

	result, err := graphdb.GetCount(rPool, gSegment, true)
	if err != nil {
		return errors.Wrapf(err, "WidgetTasks: get status count")
	}
	log.Println("result", result)
	return nil
}

func WidgetTasks(ctx context.Context, accountID, teamID string, db *sqlx.DB, rPool *redis.Pool) error {
	conditionFields := make([]graphdb.Field, 0)
	e, err := entity.RetrieveFixedEntity(ctx, db, accountID, teamID, "tasks")
	if err != nil {
		return errors.Wrapf(err, "WidgetTasks: entity retieve error")
	}

	fields, err := e.FilteredFields()
	if err != nil {
		return errors.Wrapf(err, "WidgetTasks: fields retieve error")
	}

	for _, f := range fields {
		if f.Who == entity.WhoStatus {
			conditionFields = append(conditionFields, relatable(f))
		}
	}

	//{Operator:in Key:uuid-00-contacts DataType:S Value:6eb4f58e-8327-4ccc-a262-22ad809e76cb}
	gSegment := graphdb.BuildGNode(accountID, e.ID, false).MakeBaseGNode("", conditionFields)

	result, err := graphdb.GetCount(rPool, gSegment, true)
	if err != nil {
		return errors.Wrapf(err, "WidgetTasks: get status count")
	}
	log.Println("result", result)
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
			conditionFields = append(conditionFields, makeGraphField(&f, condition.Term, condition.Expression))
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

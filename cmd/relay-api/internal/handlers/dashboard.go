package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
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
	db  *sqlx.DB
	sdb *database.SecDB
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

type GridOne struct {
	AvailableEntities    []entity.Entity      `json:"available_entities"`
	SelectedEntity       *entity.Entity       `json:"selected_entity"`
	SelectedEntityFields []entity.Field       `json:"selected_entity_fields"`
	AvailableFlows       []flow.ViewModelFlow `json:"available_flows"`
	SelectedFlow         *flow.ViewModelFlow  `json:"selected_flow"`
	Stages               []ViewModelStage     `json:"stages"`
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
	accountID := params["account_id"]
	teamID := params["team_id"]
	entityID := r.URL.Query().Get("entity_id")
	exp := r.URL.Query().Get("exp")

	gridOne := GridOne{AvailableEntities: make([]entity.Entity, 0)}
	if entityID == "" {
		entities, err := entity.AccountEntities(ctx, accountID, []int{entity.CategoryData}, d.db)
		if err != nil {
			return err
		}
		for _, e := range entities {
			if e.FlowField() != nil {
				gridOne.AvailableEntities = append(gridOne.AvailableEntities, e)
			}
		}
		if len(gridOne.AvailableEntities) > 0 {
			gridOne.SelectedEntity = &gridOne.AvailableEntities[0]
			gridOne.SelectedEntityFields = gridOne.SelectedEntity.FieldsIgnoreError()
		}
	} else {
		e, err := entity.Retrieve(ctx, accountID, entityID, d.db)
		if err != nil {
			return err
		}
		gridOne.SelectedEntity = &e
		gridOne.SelectedEntityFields = e.FieldsIgnoreError()
	}

	err := gridOne.WidgetGridOne(ctx, accountID, teamID, exp, d.db, d.sdb)
	if err != nil {
		return err
	}

	overview := struct {
		GOne GridOne `json:"g_one"`
	}{
		gridOne,
	}
	return web.Respond(ctx, w, overview, http.StatusOK)
}

func (gOne *GridOne) WidgetGridOne(ctx context.Context, accountID, teamID, exp string, db *sqlx.DB, sdb *database.SecDB) error {
	if gOne.SelectedEntity != nil && gOne.SelectedEntity.FlowField() != nil { //main stages. ex: deal stages
		flows, err := flow.List(ctx, []string{gOne.SelectedEntity.ID}, flow.FlowModePipeLine, flow.FlowTypeAll, db)
		if err != nil {
			return err
		}

		gOne.AvailableFlows = make([]flow.ViewModelFlow, len(flows))
		for i, flow := range flows {
			gOne.AvailableFlows[i] = createViewModelFlow(flow, nil)
		}

		if len(flows) > 0 {
			vmf := createViewModelFlow(flows[0], nil)
			gOne.SelectedFlow = &vmf
			gOne.Stages, err = viewModelStages(ctx, accountID, gOne.SelectedFlow.ID, db)
			if err != nil {
				return err
			}
		}
	}

	return getGoneResult(ctx, accountID, gOne, exp, db, sdb)
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
			gf := statusField.MakeGraphFieldPlain()
			if gf != nil {
				conditionFields = append(conditionFields, *gf)
			}
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
			conditionFields = append(conditionFields, f.MakeGraphField(condition.Term, condition.Expression, false))
		}
	}

	_, countMap, err := NewSegmenter(exp).
		AddCount().
		filterWrapper(ctx, params["account_id"], e.ID, fields, map[string]interface{}{}, d.db, d.sdb)
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

func otherField(key string) graphdb.Field {
	return graphdb.Field{
		Key:      key,
		DataType: graphdb.TypeString,
	}
}

func timeRange(key string, stTime, endTime time.Time) graphdb.Field {
	stInMillis := util.GetMilliSecondsStr(stTime)
	endInMillis := util.GetMilliSecondsStr(endTime)
	return graphdb.Field{
		Key:      key,
		DataType: graphdb.TypeDateRange,
		Min:      stInMillis,
		Max:      endInMillis,
		Value:    fmt.Sprintf("%s-%s", stInMillis, endInMillis),
	}
}

func sourceble(sourceEntityID string) graphdb.Field {
	return graphdb.Field{
		Value:     []interface{}{""},
		DataType:  graphdb.TypeReference,
		RefID:     sourceEntityID,
		IsReverse: false,
		Field: &graphdb.Field{
			Key:      "id",
			DataType: graphdb.TypeString,
		},
	}
}

func conditionableRef(f entity.Field, value interface{}) graphdb.Field {
	return graphdb.Field{
		Key:      f.Key,
		Value:    []interface{}{value},
		DataType: graphdb.TypeReference,
		RefID:    f.RefID,
		Field:    &graphdb.Field{},
	}
}

func (gOne *GridOne) gridResult(ctx context.Context, resultMap map[string]int, db *sqlx.DB) {
	for i := 0; i < len(gOne.Stages); i++ {
		if val, ok := resultMap[gOne.Stages[i].ID]; ok {
			gOne.Stages[i].Value = fmt.Sprintf("%d", val)
		}
	}
}

func (gThree *GridThree) gridResult(ctx context.Context, accountID, teamID string, f entity.Field, resultMap map[string]int, db *sqlx.DB) {
	e, _ := entity.Retrieve(ctx, accountID, f.RefID, db)
	refItems, _ := item.EntityItems(ctx, accountID, e.ID, db)

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

func viewModelStages(ctx context.Context, accountID, flowID string, db *sqlx.DB) ([]ViewModelStage, error) {
	nodes, err := node.NodeActorsList(ctx, accountID, flowID, db)
	if err != nil {
		return nil, err
	}

	viewModelStages := make([]ViewModelStage, 0)
	for _, n := range nodes {
		if n.Type == node.Stage {
			viewModelStages = append(viewModelStages, createViewModelStage(n))
		}
	}
	return viewModelStages, nil
}

func createViewModelStage(n node.NodeActor) ViewModelStage {
	return ViewModelStage{
		ID:          n.ID,
		FlowID:      n.FlowID,
		StageID:     n.StageID,
		Name:        n.Name,
		Description: n.Description,
		Value:       "0",
	}
}

type ViewModelStage struct {
	ID          string `json:"id"`
	FlowID      string `json:"flow_id"`
	StageID     string `json:"stage_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Value       string `json:"value"`
}

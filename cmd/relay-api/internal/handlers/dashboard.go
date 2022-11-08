package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	rg "github.com/redislabs/redisgraph-go"
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
	SelectedEntityID string               `json:"selected_entity_id"`
	SelectedFlowID   string               `json:"selected_flow_id"`
	Flows            []flow.ViewModelFlow `json:"flows"`
	Stages           []ViewModelStage     `json:"stages"`
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
	flowID := r.URL.Query().Get("flow_id")
	exp := r.URL.Query().Get("exp")

	gridOne := &GridOne{}
	e, err := entity.Retrieve(ctx, accountID, entityID, d.db, d.sdb)
	if err != nil {
		return err
	}
	gridOne.SelectedEntityID = e.ID

	if e.FlowField() == nil {
		return errors.Errorf("Cannot load grid one without flow field")
	}

	if !util.NotEmpty(flowID) {
		err = gridOne.populateFlows(ctx, accountID, teamID, d.db)
		if err != nil {
			return err
		}
	} else {
		gridOne.SelectedFlowID = flowID
	}

	err = gridOne.populateStages(ctx, accountID, teamID, d.db)
	if err != nil {
		return err
	}

	err = gridOne.itemCountPerStage(ctx, accountID, exp, e.EasyFields(), d.db, d.sdb)
	if err != nil {
		return err
	}

	err = gridOne.taskCountPerStage(ctx, accountID, teamID, d.db, d.sdb)
	if err != nil {
		return err
	}

	response := struct {
		GOne GridOne `json:"g_one"`
	}{
		*gridOne,
	}
	return web.Respond(ctx, w, response, http.StatusOK)
}

func (gOne *GridOne) populateFlows(ctx context.Context, accountID, teamID string, db *sqlx.DB) error {
	flows, err := flow.List(ctx, []string{gOne.SelectedEntityID}, flow.FlowModePipeLine, flow.FlowTypeAll, db)
	if err != nil {
		return err
	}
	gOne.Flows = make([]flow.ViewModelFlow, len(flows))
	for i, flow := range flows {
		gOne.Flows[i] = createViewModelFlow(flow, nil)
		if i == 0 {
			gOne.SelectedFlowID = flow.ID
		}
	}
	return nil
}

func (gOne *GridOne) populateStages(ctx context.Context, accountID, teamID string, db *sqlx.DB) error {
	var err error
	gOne.Stages, err = loadStages(ctx, accountID, gOne.SelectedFlowID, db)
	if err != nil {
		return err
	}
	return nil
}

func (gThree *GridThree) WidgetGridThree(ctx context.Context, accountID, teamID string, db *sqlx.DB, sdb *database.SecDB) error {
	conditionFields := make([]graphdb.Field, 0)
	e, err := entity.RetrieveFixedEntity(ctx, db, accountID, teamID, "tasks")
	if err != nil {
		return errors.Wrapf(err, "WidgetGridThree: entity retieve error")
	}

	fields := e.OnlyVisibleFields()

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
	gSegment := graphdb.BuildGNode(accountID, e.ID, false, nil).MakeBaseGNode("", conditionFields)

	result, err := graphdb.GetCount(sdb.GraphPool(), gSegment, true)
	if err != nil {
		return errors.Wrapf(err, "WidgetGridThree: get status count")
	}
	gThree.gridResult(ctx, accountID, teamID, statusField, counts(result), db, sdb)
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

	fields := e.EasyFields()
	conditionFields := make([]graphdb.Field, 0)
	whoFieldsMap := entity.WhoMap(fields)

	filter := overdue(whoFieldsMap, entity.WhoDueBy)

	for _, f := range fields {
		if condition, ok := filter.Conditions[f.Key]; ok {
			conditionFields = append(conditionFields, f.MakeGraphField(condition.Term, condition.Expression, false))
		}
	}

	_, _, err = NewSegmenter(exp).
		AddCount().
		filterWrapper(ctx, params["account_id"], e.ID, fields, map[string]interface{}{}, d.db, d.sdb)
	if err != nil {
		return err
	}

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

func sourcebleItem(sourceEntityID, sourceItemID string) graphdb.Field {
	return graphdb.Field{
		Value:     []interface{}{sourceItemID},
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

func publicRecordsOnly() graphdb.Field {
	return graphdb.Field{
		Key:        "system_is_public",
		Value:      true,
		DataType:   graphdb.TypeString,
		IsReverse:  false,
		Expression: "=",
	}
}

func (gOne *GridOne) gridResult(ctx context.Context, resultMap map[string]int, db *sqlx.DB) {
	for i := 0; i < len(gOne.Stages); i++ {
		if val, ok := resultMap[gOne.Stages[i].ID]; ok {
			gOne.Stages[i].Value = fmt.Sprintf("%d", val)
		}
	}
}

func (gOne *GridOne) taskCountForEachStage(ctx context.Context, statusField entity.Field, result *rg.QueryResult, db *sqlx.DB) {
	response := make(map[string][]Series, 0)
	if result == nil {
		return
	}
	for result.Next() { // Next returns true until the iterator is depleted.
		// Get the current Record.

		r := result.Record()

		stageID := r.GetByIndex(1).(string)
		statusID := r.GetByIndex(2).(string)

		choice := statusField.ChoiceMap()[statusID]
		if _, ok := response[stageID]; !ok {
			response[stageID] = make([]Series, 0)
		}

		switch v := r.GetByIndex(0).(type) {
		case int:
			response[stageID] = append(response[stageID], createPartialVMSeries(choice.ID, choice.DisplayValue.(string), choice.Color, choice.Verb, v))
		case float64:
			response[stageID] = append(response[stageID], createPartialVMSeries(choice.ID, choice.DisplayValue.(string), choice.Color, choice.Verb, int(v)))
		}
	}

	for i := 0; i < len(gOne.Stages); i++ {
		gOne.Stages[i].Series = response[gOne.Stages[i].ID]
	}
}

func (gThree *GridThree) gridResult(ctx context.Context, accountID, teamID string, f entity.Field, resultMap map[string]int, db *sqlx.DB, sdb *database.SecDB) {
	e, _ := entity.Retrieve(ctx, accountID, f.RefID, db, sdb)
	refItems, _ := item.EntityItems(ctx, accountID, e.ID, db)

	choicer := reference.ItemChoices(&f, refItems, e.WhoKeyMap())
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

func loadStages(ctx context.Context, accountID, flowID string, db *sqlx.DB) ([]ViewModelStage, error) {
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
	ID          string   `json:"id"`
	FlowID      string   `json:"flow_id"`
	StageID     string   `json:"stage_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Value       string   `json:"value"`
	Series      []Series `json:"series"`
}

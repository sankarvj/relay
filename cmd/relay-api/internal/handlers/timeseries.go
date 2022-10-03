package handlers

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/redislabs/redisgraph-go"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/event"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/stream"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/reference"
	"gitlab.com/vjsideprojects/relay/internal/timeseries"
	"go.opencensus.io/trace"
)

type Timeseries struct {
	db            *sqlx.DB
	sdb           *database.SecDB
	authenticator *auth.Authenticator
}

// Create inserts a new team into the system.
func (ts *Timeseries) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	log := log.New(os.Stdout, "RELAY TIMESERIES : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	ctx, span := trace.StartSpan(ctx, "handlers.Timeseries.Create")
	defer span.End()

	var ne event.NewEvent
	if err := web.Decode(r, &ne); err != nil {
		return errors.Wrap(err, "")
	}
	accountID := params["account_id"]
	entityName := strValue(ne.Body["module"])

	tsData, err := event.Process(ctx, accountID, entityName, ne.Body, log, ts.db)
	if err != nil {
		return errors.Wrapf(err, "process failed")
	}

	if tsData.OldData == nil && tsData.NewData != nil {
		log.Println("processEvent : started : sqs streaming")
		err = job.NewJob(ts.db, ts.sdb, ts.authenticator.FireBaseAdminSDK).Stream(stream.NewEventItemMessage(ctx, ts.db, accountID, "", tsData.NewData.EntityID, tsData.NewData.ID, tsData.NewData.Fields(), nil))
		if err != nil {
			log.Println("processEvent : errored : sqs streaming", err)
		}
		log.Println("processEvent : completed : sqs streaming")
	} else if tsData.OldData != nil && tsData.NewData != nil {
		log.Println("processEvent : started : sqs streaming")
		err = job.NewJob(ts.db, ts.sdb, ts.authenticator.FireBaseAdminSDK).Stream(stream.NewEventItemMessage(ctx, ts.db, accountID, "", tsData.NewData.EntityID, tsData.NewData.ID, tsData.NewData.Fields(), tsData.OldData.Fields()))
		if err != nil {
			log.Println("processEvent : errored : sqs streaming", err)
		}
		log.Println("processEvent : completed : sqs streaming")
	}

	return web.Respond(ctx, w, tsData.NewData, http.StatusCreated)
}

func (ts *Timeseries) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Stream.List")
	defer span.End()

	accountID, entityID, _ := takeAEI(ctx, params, ts.db)
	startTime, endTime := timeseries.Duration(r.URL.Query().Get("duration"))

	series, err := timeseries.List(ctx, accountID, entityID, startTime, endTime, ts.db)
	if err != nil {
		return err
	}

	ch := Chart{
		Series: vmseries(series),
		Title:  "Daily active users",
		Type:   "line",
	}

	return web.Respond(ctx, w, ch, http.StatusOK)
}

func (ts *Timeseries) Count(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	var countByField entity.Field
	countByFieldName := params["count_by"]
	accountID, entityID, _ := takeAEI(ctx, params, ts.db)
	e, err := entity.Retrieve(ctx, accountID, entityID, ts.db)
	if err != nil {
		return err
	}

	fields, err := e.FilteredFields()
	if err != nil {
		return errors.Wrapf(err, "WidgetGridThree: fields retieve error")
	}

	exp := r.URL.Query().Get("exp")
	grouped, _ := strconv.ParseBool((r.URL.Query().Get("grouped")))
	conditionFields, err := makeConditionsFromExp(ctx, accountID, entityID, exp, ts.db, ts.sdb)
	if err != nil {
		return err
	}
	for _, f := range fields {
		if f.Name == countByFieldName {
			countByField = f
			if countByField.IsReference() {
				conditionFields = append(conditionFields, relatable(f))
			} else if countByField.IsList() {
				conditionFields = append(conditionFields, listable(f))
			}
		}
	}

	if countByField.IsReference() {
		refItems, err := item.EntityItems(ctx, accountID, e.ID, ts.db)
		if err != nil {
			log.Printf("***> unexpected error occurred when retriving reference items for field unit inside updating choices error: %v.\n continuing...", err)
		}
		reference.ChoicesMaker(&countByField, "", reference.ItemChoices(&countByField, refItems, e.WhoFields()))
	}

	//{Operator:in Key:uuid-00-contacts DataType:S Value:6eb4f58e-8327-4ccc-a262-22ad809e76cb}
	gSegment := graphdb.BuildGNode(accountID, e.ID, false).MakeBaseGNode("", conditionFields)

	var result *redisgraph.QueryResult
	if !grouped {
		result, err = graphdb.GetCount(ts.sdb.GraphPool(), gSegment, true)
	} else {
		result, err = graphdb.GetGroupedCount(ts.sdb.GraphPool(), gSegment, countByField.Key)
	}

	if err != nil {
		return errors.Wrapf(err, "WidgetGridThree: get status count")
	}
	cr := counts(result)

	ch := Chart{
		Series: vmseriesFromMap(cr, countByField),
		Title:  "Accounts by status",
		Type:   "bar",
	}

	return web.Respond(ctx, w, ch, http.StatusOK)
}

func (ts *Timeseries) Sum(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	var sumByField entity.Field
	sumByFieldName := params["sum_by"]
	duration := r.URL.Query().Get("duration")
	accountID, entityID, _ := takeAEI(ctx, params, ts.db)
	e, err := entity.Retrieve(ctx, accountID, entityID, ts.db)
	if err != nil {
		return err
	}

	fields, err := e.FilteredFields()
	if err != nil {
		return errors.Wrapf(err, "WidgetGridThree: fields retieve error")
	}

	conditionFields := make([]graphdb.Field, 0)
	for _, f := range fields {
		if f.Name == sumByFieldName {
			sumByField = f
		}
	}
	conditionFields = append(conditionFields, timeRange("system_created_at", duration))

	//{Operator:in Key:uuid-00-contacts DataType:S Value:6eb4f58e-8327-4ccc-a262-22ad809e76cb}
	gSegment := graphdb.BuildGNode(accountID, e.ID, false).MakeBaseGNode("", conditionFields)

	result, err := graphdb.GetSum(ts.sdb.GraphPool(), gSegment, sumByField.Key)
	if err != nil {
		return errors.Wrapf(err, "WidgetGridThree: get status sum")
	}
	cr := counts(result)

	if sumByField.IsReference() {
		refItems, err := item.EntityItems(ctx, accountID, e.ID, ts.db)
		if err != nil {
			log.Printf("***> unexpected error occurred when retriving reference items for field unit inside updating choices error: %v.\n continuing...", err)
		}
		reference.ChoicesMaker(&sumByField, "", reference.ItemChoices(&sumByField, refItems, e.WhoFields()))
	}
	ch := Chart{
		Series: vmseriesFromMap(cr, sumByField),
		Title:  "Sum of NPS",
		Type:   "pie",
	}

	return web.Respond(ctx, w, ch, http.StatusOK)
}

func (ts *Timeseries) GroupByCount(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	var countByField entity.Field
	countByFieldName := params["count_by"]
	accountID, entityID, _ := takeAEI(ctx, params, ts.db)
	e, err := entity.Retrieve(ctx, accountID, entityID, ts.db)
	if err != nil {
		return err
	}

	fields, err := e.FilteredFields()
	if err != nil {
		return errors.Wrapf(err, "WidgetGridThree: fields retieve error")
	}

	conditionFields := make([]graphdb.Field, 0)
	for _, f := range fields {
		if f.Name == countByFieldName {
			countByField = f
			conditionFields = append(conditionFields, otherField(f.Key))
		}
	}

	if countByField.IsReference() {
		refItems, err := item.EntityItems(ctx, accountID, e.ID, ts.db)
		if err != nil {
			log.Printf("***> unexpected error occurred when retriving reference items for field unit inside updating choices error: %v.\n continuing...", err)
		}
		reference.ChoicesMaker(&countByField, "", reference.ItemChoices(&countByField, refItems, e.WhoFields()))
	}

	//{Operator:in Key:uuid-00-contacts DataType:S Value:6eb4f58e-8327-4ccc-a262-22ad809e76cb}
	gSegment := graphdb.BuildGNode(accountID, e.ID, false).MakeBaseGNode("", conditionFields)

	result, err := graphdb.GetCount(ts.sdb.GraphPool(), gSegment, true)
	if err != nil {
		return errors.Wrapf(err, "WidgetGridThree: get status count")
	}
	cr := counts(result)

	ch := Chart{
		Series: vmseriesFromMap(cr, countByField),
		Title:  "Accounts by status",
		Type:   "bar",
	}

	return web.Respond(ctx, w, ch, http.StatusOK)
}

func vmseries(tms []timeseries.Timeseries) []Series {
	vmseries := make([]Series, len(tms))
	for i, ts := range tms {
		vmseries[i] = createVMSeries(ts)
	}
	return vmseries
}

func vmseriesFromMap(m map[string]int, f entity.Field) []Series {
	mapOfChoices := f.ChoicesMap()

	vmseries := make([]Series, 0)
	for label, value := range m {
		if val, ok := mapOfChoices[label]; ok {
			label = util.ConvertIntfToStr(val.DisplayValue)
		}
		vmseries = append(vmseries, createVMSeriesFromMap(label, value))
	}
	return vmseries
}

func strValue(v interface{}) string {
	if v != nil {
		return v.(string)
	}
	return ""
}

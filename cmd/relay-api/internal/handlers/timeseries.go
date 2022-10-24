package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/chart"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/event"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/mid"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/stream"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
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
	ctx, span := trace.StartSpan(ctx, "handlers.Timeseries.List")
	defer span.End()

	accountID := params["account_id"]
	entityID := params["entity_id"]
	zone, _ := util.ParseTime(r.URL.Query().Get("zone"))
	loc := time.FixedZone(zone.Zone())
	exp := r.URL.Query().Get("exp")
	baseEntityID := r.URL.Query().Get("be")
	//baseItemID := r.URL.Query().Get("bi")

	charts, err := chart.List(ctx, accountID, baseEntityID, entityID, ts.db)
	if err != nil {
		return err
	}

	gridResMap, err := grids(ctx, charts, exp, loc, ts.db, ts.sdb)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, createViewModelCharts(charts, gridResMap), http.StatusOK)
}

func (ts *Timeseries) Chart(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Timeseries.Chart")
	defer span.End()

	exp := r.URL.Query().Get("exp")
	duration := r.URL.Query().Get("duration")
	zone, _ := util.ParseTime(r.URL.Query().Get("zone"))
	baseEntityID := r.URL.Query().Get("be")
	baseItemID := r.URL.Query().Get("bi")

	ch, err := chart.Retrieve(ctx, params["account_id"], params["chart_id"], ts.db)
	if err != nil {
		return err
	}
	if util.NotEmpty(baseEntityID) && util.NotEmpty(baseItemID) {
		//TODO remove middleware logic from here....
		//checking this here because we are using the ch.BaseEntityID directly inside list
		if baseEntityID != ch.BaseEntityID {
			return mid.ErrForbidden
		}
	}
	if duration != "undefined" && duration != "" {
		ch.Duration = duration
	}
	startTime, endTime, lastStart := timeseries.DurationWithZone(ch.Duration, time.FixedZone(zone.Zone()))

	var vmc VMChart
	var series []timeseries.Timeseries
	switch ch.GetDType() {
	case string(chart.DTypeTimeseries):
		series, err = timeseries.List(ctx, ch.AccountID, ch.EntityID, startTime, endTime, ts.db)
		count, err := timeseries.Count(ctx, ch.AccountID, ch.EntityID, lastStart, startTime, ts.db)
		if err != nil {
			return err
		}
		vmc = createViewModelChart(*ch, vmseries(series), len(series), change(len(series), count))
	case string(chart.DTypeDefault):
		series, err := list(ctx, *ch, exp, baseItemID, startTime, endTime, ts.db, ts.sdb)
		if err != nil {
			return err
		}
		vmc = createViewModelChartNoChange(*ch, series, 0)
	}
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, vmc, http.StatusOK)
}

func (ts *Timeseries) Overview(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	//duration := r.URL.Query().Get("duration")
	exp := r.URL.Query().Get("exp")
	accountID, entityID, _ := takeAEI(ctx, params, ts.db)
	charts, err := chart.List(ctx, accountID, entity.NoEntityID, entityID, ts.db)
	if err != nil {
		return err
	}

	vmCharts := make([]VMChart, 0)
	for _, ch := range charts {
		stTime, endTime, _ := timeseries.Duration(ch.Duration)

		series, err := list(ctx, ch, exp, "", stTime, endTime, ts.db, ts.sdb)
		if err != nil {
			return err
		}
		vmc := createViewModelChartNoChange(ch, series, 0)
		vmCharts = append(vmCharts, vmc)
	}

	return web.Respond(ctx, w, vmCharts, http.StatusOK)
}

//TODO Worst non-generic way of implementation...rewrite this block
func (ts *Timeseries) CSMOverview(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	//duration := r.URL.Query().Get("duration")
	exp := r.URL.Query().Get("exp")
	accountID, entityID, _ := takeAEI(ctx, params, ts.db)
	ch, err := chart.Retrieve(ctx, accountID, params["chart_id"], ts.db)
	if err != nil {
		return err
	}
	//remove middleware logic from here....
	if entityID != ch.EntityID {
		return mid.ErrForbidden
	}

	e, err := entity.Retrieve(ctx, accountID, entityID, ts.db, ts.sdb)
	if err != nil {
		return err
	}

	baseEntityID := r.URL.Query().Get("be")
	baseEntity, err := entity.Retrieve(ctx, accountID, baseEntityID, ts.db, ts.sdb)
	if err != nil {
		return err
	}
	baseItemID := r.URL.Query().Get("bi")
	baseItem, err := item.Retrieve(ctx, baseEntityID, baseItemID, ts.db)
	if err != nil {
		return err
	}
	assCompanies := baseEntity.Key("associated_companies")
	val := baseItem.Fields()[assCompanies]

	stTime, endTime, _ := timeseries.Duration(ch.Duration)

	exp1 := fmt.Sprintf("{{%s.%s}} in {%s}", entityID, e.Key("associated_companies"), util.ConvertIntfToCommaSepString(val))
	exp = util.AddExpression(exp, exp1)
	log.Println("exp ", exp)
	series, err := list(ctx, *ch, exp, "", stTime, endTime, ts.db, ts.sdb)
	if err != nil {
		return err
	}
	log.Println("series ", series)
	vmc := createViewModelChartNoChange(*ch, series, 0)

	return web.Respond(ctx, w, vmc, http.StatusOK)
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
	for id, value := range m {
		label := id // this line fixes for group with name
		if val, ok := mapOfChoices[id]; ok {
			label = util.ConvertIntfToStr(val.DisplayValue)
		}
		vmseries = append(vmseries, createVMSeriesFromMap(id, label, value))
	}
	return vmseries
}

func strValue(v interface{}) string {
	if v != nil {
		return v.(string)
	}
	return ""
}

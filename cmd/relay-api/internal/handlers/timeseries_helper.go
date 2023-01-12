package handlers

import (
	"context"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/account"
	"gitlab.com/vjsideprojects/relay/internal/chart"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/database/dbservice"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/reference"
	"gitlab.com/vjsideprojects/relay/internal/timeseries"
)

func loadSeries(ctx context.Context, ch chart.Chart, exp string, stTime, endTime time.Time, db *sqlx.DB, sdb *database.SecDB) ([]Series, error) {
	return loadCHSeries(ctx, ch, exp, NoEntityID, NoEntityID, stTime, endTime, db, sdb)
}

func loadCHSeries(ctx context.Context, ch chart.Chart, exp, baseEntityID, baseItemID string, stTime, endTime time.Time, db *sqlx.DB, sdb *database.SecDB) ([]Series, error) {
	e, err := entity.Retrieve(ctx, ch.AccountID, ch.EntityID, db, sdb)
	if err != nil {
		return nil, err
	}
	fields := e.EasyFields()

	source := ch.GetSource()
	fieldName := ch.GetField()
	fieldDate := ch.GetDate()
	groupedLogic := ch.GetGroupByLogic()

	conditionFields, err := makeConditionsFromExp(ctx, ch.AccountID, ch.EntityID, exp, db, sdb)
	if err != nil {
		return nil, err
	}
	//add source condition if source exists
	if source != entity.NoEntityID && source != ch.EntityID {
		conditionFields = append(conditionFields, sourceble(source))
	}

	//must add base condition if base exists - very important.
	if util.NotEmpty(baseEntityID) && util.NotEmpty(baseItemID) {
		conditionFields = append(conditionFields, sourcebleItem(baseEntityID, baseItemID))
	}

	var filterByField entity.Field
	for _, f := range fields {
		if f.Name == fieldName {
			filterByField = f

			gf := filterByField.MakeGraphFieldPlain()
			if gf != nil {
				conditionFields = append(conditionFields, *gf)
			}

			// populate field choices
			if f.IsReference() && (!f.IsNode() || baseEntityID == NoEntityID || baseEntityID == "undefined") {
				refItems, err := item.EntityItems(ctx, ch.AccountID, filterByField.RefID, db)
				if err != nil {
					log.Printf("***> unexpected error occurred when retriving reference items for field unit inside updating choices error: %v.\n continuing...", err)
				}
				reference.ChoicesMaker(&filterByField, "", reference.ItemChoices(&filterByField, refItems, map[string]string{}))
			} else if f.IsNode() {
				nodes, err := loadNodes(ctx, ch.AccountID, baseEntityID, baseItemID, db, sdb)
				if err != nil {
					return nil, err
				}
				reference.ChoicesMaker(&filterByField, "", reference.NodeActorChoices(nodes))
			}
		} else if f.Name == fieldDate {
			conditionFields = append(conditionFields, timeRange(f.Key, stTime, endTime))
		}
	}

	// log.Printf("chart ---:: %+v", ch)
	// log.Printf("chart conditionFields---%+v", conditionFields)

	useDB := account.UseDB(ctx, db, ch.AccountID)
	var counters []dbservice.Counters
	switch groupedLogic {
	case string(chart.GroupLogicID):
		counters, err = dbservice.NewDBservice(useDB, db, sdb).Count(ctx, ch.AccountID, e.ID, "", "", chart.GroupLogicID, conditionFields)
	case string(chart.GroupLogicField):
		counters, err = dbservice.NewDBservice(useDB, db, sdb).Count(ctx, ch.AccountID, e.ID, filterByField.Key, filterByField.RefID, chart.GroupLogicField, conditionFields)
	case string(chart.GroupLogicParent):
		counters, err = dbservice.NewDBservice(useDB, db, sdb).Count(ctx, ch.AccountID, e.ID, "", "", chart.GroupLogicParent, conditionFields)
	case string(chart.GroupLogicNone):
		counters, err = dbservice.NewDBservice(useDB, db, sdb).Count(ctx, ch.AccountID, e.ID, "", "", chart.GroupLogicNone, conditionFields)
	}

	if err != nil {
		return nil, err
	}

	return vmseriesFromMap(counts(counters), filterByField), nil
}

func sum(ctx context.Context, ch chart.Chart, exp string, db *sqlx.DB, sdb *database.SecDB) ([]Series, error) {
	e, err := entity.Retrieve(ctx, ch.AccountID, ch.EntityID, db, sdb)
	if err != nil {
		return nil, err
	}
	fields := e.EasyFields()
	source := ch.GetSource()
	//groupedLogic := ch.GetGroupByLogic()
	fieldName := ch.GetField()
	startTime, endTime, _ := timeseries.Duration(ch.Duration)

	conditionFields, err := makeConditionsFromExp(ctx, ch.AccountID, ch.EntityID, exp, db, sdb)
	if err != nil {
		return nil, err
	}
	//add source condition if source exists
	if source != entity.NoEntityID && source != ch.EntityID {
		conditionFields = append(conditionFields, sourceble(source))
	}

	var filterByField entity.Field
	for _, f := range fields {
		if f.Name == fieldName {
			filterByField = f
		}
	}
	if filterByField.IsReference() {
		refItems, err := item.EntityItems(ctx, ch.AccountID, filterByField.RefID, db)
		if err != nil {
			log.Printf("***> unexpected error occurred when retriving reference items for field unit inside updating choices error: %v.\n continuing...", err)
		}
		reference.ChoicesMaker(&filterByField, "", reference.ItemChoices(&filterByField, refItems, e.WhoKeyMap()))
	}

	conditionFields = append(conditionFields, timeRange("system_created_at", startTime, endTime))
	summers, err := dbservice.NewDBservice(dbservice.Spider, db, sdb).Sum(ctx, ch.AccountID, ch.EntityID, filterByField.Key, conditionFields)
	if err != nil {
		return nil, err
	}

	return vmseriesFromMap(counts(summers), filterByField), nil
}

func grids(ctx context.Context, charts []chart.Chart, exp string, loc *time.Location, db *sqlx.DB, sdb *database.SecDB) (map[string]EagerLoader, error) {
	gridResMap := make(map[string]EagerLoader, 0)
	var err error
	for _, ch := range charts {
		if ch.Type == string(chart.TypeGrid) {
			var newcount, oldcount int
			switch ch.GetDType() {
			case string(chart.DTypeDefault):
				newcount, oldcount, err = gridDefault(ctx, ch.AccountID, ch.EntityID, ch.Duration, ch, loc, db, sdb)
				if err != nil {
					return nil, err
				}
			case string(chart.DTypeTimeseries):
				newcount, oldcount, err = gridTimeseries(ctx, ch.AccountID, ch.EntityID, ch.Duration, loc, db)
				if err != nil {
					return nil, err
				}
			}
			gridResMap[ch.ID] = EagerLoader{
				Count:  newcount,
				Change: change(newcount, oldcount),
				Series: []Series{},
			}
		}
	}

	return gridResMap, nil
}

func gridTimeseries(ctx context.Context, accountID, entityID, duration string, loc *time.Location, db *sqlx.DB) (int, int, error) {
	startTime, endTime, lastStart := timeseries.DurationWithZone(duration, loc)

	countNew, err := timeseries.Count(ctx, accountID, entityID, startTime, endTime, db)
	if err != nil {
		return 0, 0, err
	}
	countOld, err := timeseries.Count(ctx, accountID, entityID, lastStart, startTime, db)
	if err != nil {
		return 0, 0, err
	}
	return countNew, countOld, nil
}

func gridDefault(ctx context.Context, accountID, entityID, duration string, ch chart.Chart, loc *time.Location, db *sqlx.DB, sdb *database.SecDB) (int, int, error) {
	stTime, endTime, lastStart := timeseries.DurationWithZone(duration, loc)
	series, err := loadSeries(ctx, ch, ch.GetExp(), stTime, endTime, db, sdb)
	if err != nil {
		return 0, 0, err
	}
	switch ch.GetCalc() {
	case string(chart.CalcSum):
		if len(series) > 0 {
			return series[0].Count, 0, nil
		}
	case string(chart.CalcCount):
		if len(series) > 0 {
			return series[0].Count, 0, nil
		}
	case string(chart.CalcRate):
		oldseries, err := loadSeries(ctx, ch, ch.GetExp(), lastStart, stTime, db, sdb)
		if err != nil {
			return 0, 0, err
		}
		if len(series) > 0 && len(oldseries) > 0 {
			return series[0].Count, oldseries[0].Count, nil
		} else if len(series) > 0 {
			return series[0].Count, 0, nil
		} else if len(oldseries) > 0 {
			return 0, oldseries[0].Count, nil
		}
	}
	return 0, 0, nil
}

func change(count, oldCount int) int {
	total := count + oldCount
	if total == 0 {
		return 0
	}
	return (count / total) * 100
}

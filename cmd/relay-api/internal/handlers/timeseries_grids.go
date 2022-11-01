package handlers

import (
	"context"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redislabs/redisgraph-go"
	"gitlab.com/vjsideprojects/relay/internal/chart"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/reference"
	"gitlab.com/vjsideprojects/relay/internal/timeseries"
)

func list(ctx context.Context, ch chart.Chart, exp string, stTime, endTime time.Time, db *sqlx.DB, sdb *database.SecDB) ([]Series, error) {
	return listwb(ctx, ch, exp, NoEntityID, NoEntityID, stTime, endTime, db, sdb)
}

func listwb(ctx context.Context, ch chart.Chart, exp, baseEntityID, baseItemID string, stTime, endTime time.Time, db *sqlx.DB, sdb *database.SecDB) ([]Series, error) {
	e, err := entity.Retrieve(ctx, ch.AccountID, ch.EntityID, db, sdb)
	if err != nil {
		return nil, err
	}
	fields := e.FieldsIgnoreError()

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
			if f.IsReference() {
				refItems, err := item.EntityItems(ctx, ch.AccountID, filterByField.RefID, db)
				if err != nil {
					log.Printf("***> unexpected error occurred when retriving reference items for field unit inside updating choices error: %v.\n continuing...", err)
				}
				reference.ChoicesMaker(&filterByField, "", reference.ItemChoices(&filterByField, refItems, map[string]string{}))
			}
		} else if f.Name == fieldDate {
			conditionFields = append(conditionFields, timeRange(f.Key, stTime, endTime))
		}
	}

	//{Operator:in Key:uuid-00-contacts DataType:S Value:6eb4f58e-8327-4ccc-a262-22ad809e76cb}
	gSegment := graphdb.BuildGNode(ch.AccountID, e.ID, false, nil).MakeBaseGNode("", conditionFields)
	var result *redisgraph.QueryResult
	switch groupedLogic {
	case string(chart.GroupLogicID):
		result, err = graphdb.GetCount(sdb.GraphPool(), gSegment, true)
	case string(chart.GroupLogicField):
		result, err = graphdb.GetGroupedCount(sdb.GraphPool(), gSegment, filterByField.Key)
	case string(chart.GroupLogicParent):
		result, err = graphdb.GetFromParentCount(sdb.GraphPool(), gSegment)
	case string(chart.GroupLogicNone):
		result, err = graphdb.GetCount(sdb.GraphPool(), gSegment, false)
	}

	if err != nil {
		return nil, err
	}

	return vmseriesFromMap(counts(result), filterByField), nil
}

func sum(ctx context.Context, ch chart.Chart, exp string, db *sqlx.DB, sdb *database.SecDB) ([]Series, error) {
	e, err := entity.Retrieve(ctx, ch.AccountID, ch.EntityID, db, sdb)
	if err != nil {
		return nil, err
	}
	fields := e.FieldsIgnoreError()
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
		reference.ChoicesMaker(&filterByField, "", reference.ItemChoices(&filterByField, refItems, e.WhoFields()))
	}

	conditionFields = append(conditionFields, timeRange("system_created_at", startTime, endTime))
	//{Operator:in Key:uuid-00-contacts DataType:S Value:6eb4f58e-8327-4ccc-a262-22ad809e76cb}
	gSegment := graphdb.BuildGNode(ch.AccountID, ch.EntityID, false, nil).MakeBaseGNode("", conditionFields)
	result, err := graphdb.GetSum(sdb.GraphPool(), gSegment, filterByField.Key)
	if err != nil {
		return nil, err
	}

	return vmseriesFromMap(counts(result), filterByField), nil
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
	series, err := list(ctx, ch, ch.GetExp(), stTime, endTime, db, sdb)
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
		oldseries, err := list(ctx, ch, ch.GetExp(), lastStart, stTime, db, sdb)
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

// func sss(){
// 	newC, oldC, err := gridDAU(ctx, ch.AccountID, entity.FixedEntityDailyActiveUsers, "last_24hrs", ts.db)
// 	if err != nil {
// 		return err
// 	}

// 	ch1 := VMChart{
// 		Title: "DAU",
// 		Type:  "grid",
// 		Count: change(newC, oldC),
// 	}

// 	lost, total, err := gridChrun(ctx, params["account_id"], params["team_id"], entity.FixedEntityContacts, "lifecycle_stage", "became_a_customer_date", "1", "last_24hrs", ts.db, ts.sdb)
// 	if err != nil {
// 		return err
// 	}

// 	ch2 := VMChart{
// 		Title: "Churn Rate",
// 		Type:  "grid",
// 		Count: rate(lost, total),
// 	}

// 	newcustomers, err := gridNewCustomers(ctx, params["account_id"], params["team_id"], entity.FixedEntityContacts, "", "became_a_customer_date", nil, "last_24hrs", ts.db, ts.sdb)
// 	if err != nil {
// 		return err
// 	}

// 	ch3 := VMChart{
// 		Title: "New Customer",
// 		Type:  "grid",
// 		Count: newcustomers,
// 	}

// 	dacc, err := delayedAccounts(ctx, params["account_id"], params["team_id"], entity.FixedEntityProjects, "end_time", entity.FixedEntityCompanies, "", ts.db, ts.sdb)
// 	if err != nil {
// 		return err
// 	}

// 	ch4 := VMChart{
// 		Title: "Accounts with delayed projects",
// 		Type:  "grid",
// 		Count: dacc,
// 	}
// }

// func gridChrun(ctx context.Context, accountID, teamID, entityName, fieldName, fieldDate string, fieldValue interface{}, duration string, db *sqlx.DB, sdb *database.SecDB) (int, int, error) {
// 	startTime, endTime, _ := timeseries.Duration(duration)
// 	matched1, _, err := churn(ctx, accountID, teamID, entityName, fieldName, fieldDate, fieldValue, startTime, endTime, db, sdb)
// 	if err != nil {
// 		return 0, 0, err
// 	}
// 	difference := startTime.Sub(endTime)
// 	newStart := startTime.Add(difference)
// 	_, allOtherCount2, err := churn(ctx, accountID, teamID, entityName, fieldName, fieldDate, fieldValue, newStart, startTime, db, sdb)
// 	if err != nil {
// 		return 0, 0, err
// 	}
// 	return matched1, allOtherCount2, nil
// }

// func gridNewCustomers(ctx context.Context, accountID, teamID, entityName, fieldName, fieldDate string, fieldValue interface{}, duration string, db *sqlx.DB, sdb *database.SecDB) (int, error) {
// 	startTime, endTime, _ := timeseries.Duration(duration)
// 	e, err := entity.RetrieveFixedEntity(ctx, db, accountID, teamID, entityName)
// 	if err != nil {
// 		return 0, err
// 	}
// 	fields := e.FieldsIgnoreError()

// 	conditionFields := make([]graphdb.Field, 0)
// 	for _, f := range fields {
// 		if f.Name == fieldName {
// 			conditionFields = append(conditionFields, *f.MakeGraphFieldPlain())
// 		}

// 		if f.Name == fieldDate {
// 			conditionFields = append(conditionFields, timeRange(f.Key, startTime, endTime))
// 		}
// 	}

// 	gSegment := graphdb.BuildGNode(accountID, e.ID, false).MakeBaseGNode("", conditionFields)
// 	var result *redisgraph.QueryResult
// 	result, err = graphdb.GetCount(sdb.GraphPool(), gSegment, false)
// 	if err != nil {
// 		return 0, err
// 	}
// 	cr := counts(result)
// 	return cr["total_count"], nil

// }

// func churn(ctx context.Context, accountID, teamID, entityName, fieldName, fieldDate string, fieldValue interface{}, startTime, endTime time.Time, db *sqlx.DB, sdb *database.SecDB) (int, int, error) {
// 	e, err := entity.RetrieveFixedEntity(ctx, db, accountID, teamID, entityName)
// 	if err != nil {
// 		return 0, 0, err
// 	}
// 	fields := e.FieldsIgnoreError()

// 	conditionFields := make([]graphdb.Field, 0)
// 	for _, f := range fields {
// 		if f.Name == fieldName {
// 			conditionFields = append(conditionFields, *f.MakeGraphFieldPlain())
// 		}

// 		if f.Name == fieldDate {
// 			conditionFields = append(conditionFields, timeRange(f.Key, startTime, endTime))
// 		}
// 	}

// 	gSegment := graphdb.BuildGNode(accountID, e.ID, false).MakeBaseGNode("", conditionFields)
// 	var result *redisgraph.QueryResult
// 	result, err = graphdb.GetCount(sdb.GraphPool(), gSegment, true)
// 	if err != nil {
// 		return 0, 0, err
// 	}
// 	cr := counts(result)

// 	matched, allOther := allButMatched(fieldValue, cr)
// 	return matched, allOther, nil
// }

// func delayedAccounts(ctx context.Context, accountID, teamID, entityName, dateFieldName, sourceEntityName, exp string, db *sqlx.DB, sdb *database.SecDB) (int, error) {
// 	se, err := entity.RetrieveFixedEntity(ctx, db, accountID, teamID, sourceEntityName)
// 	if err != nil {
// 		return 0, err
// 	}

// 	e, err := entity.RetrieveFixedEntity(ctx, db, accountID, teamID, entityName)
// 	if err != nil {
// 		return 0, err
// 	}
// 	fields := e.FieldsIgnoreError()

// 	conditionFields, err := makeConditionsFromExp(ctx, accountID, e.ID, exp, db, sdb)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if se.ID != "" {
// 		conditionFields = append(conditionFields, sourceble(se.ID))
// 	}
// 	for _, f := range fields {
// 		if f.Name == dateFieldName {
// 			dateexp := fmt.Sprintf("{{%s.%s}} bf {%s}", e.ID, f.Key, "now")
// 			dateConditions, _ := makeConditionsFromExp(ctx, accountID, e.ID, dateexp, db, sdb)
// 			conditionFields = append(conditionFields, dateConditions...)
// 		}
// 	}

// 	gSegment := graphdb.BuildGNode(accountID, e.ID, false).MakeBaseGNode("", conditionFields)
// 	result, err := graphdb.GetFromParentCount(sdb.GraphPool(), gSegment)
// 	if err != nil {
// 		return 0, err
// 	}
// 	crAffected := counts(result)

// 	return crAffected["total_count"], nil
// }

func allButMatched(fieldValue interface{}, counts map[string]int) (int, int) {
	var matchedCount int
	var allOthersCount int
	for id, count := range counts {
		if id == fieldValue.(string) {
			matchedCount = matchedCount + count
		} else {
			allOthersCount = allOthersCount + count
		}
	}
	return matchedCount, allOthersCount
}

func change(count, oldCount int) int {
	total := count + oldCount
	if total == 0 {
		return 0
	}
	return (count / total) * 100
}

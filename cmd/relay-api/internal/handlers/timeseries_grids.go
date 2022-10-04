package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redislabs/redisgraph-go"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/timeseries"
)

func gridDAU(ctx context.Context, accountID, entityName, duration string, db *sqlx.DB) (int, int, error) {
	startTime, endTime := timeseries.Duration(duration)
	difference := startTime.Sub(endTime)
	newStart := startTime.Add(difference)
	e, err := entity.RetrieveFixedEntityAccountLevel(ctx, db, accountID, entityName)
	if err != nil {
		return 0, 0, err
	}
	countNew, err := timeseries.Count(ctx, accountID, e.ID, startTime, endTime, db)
	if err != nil {
		return 0, 0, err
	}
	countOld, err := timeseries.Count(ctx, accountID, e.ID, newStart, startTime, db)
	if err != nil {
		return 0, 0, err
	}
	return countNew, countOld, nil
}

func gridChrun(ctx context.Context, accountID, teamID, entityName, fieldName, fieldDate string, fieldValue interface{}, duration string, db *sqlx.DB, sdb *database.SecDB) (int, int, error) {
	startTime, endTime := timeseries.Duration(duration)
	matched1, _, err := churn(ctx, accountID, teamID, entityName, fieldName, fieldDate, fieldValue, startTime, endTime, db, sdb)
	if err != nil {
		return 0, 0, err
	}
	difference := startTime.Sub(endTime)
	newStart := startTime.Add(difference)
	_, allOtherCount2, err := churn(ctx, accountID, teamID, entityName, fieldName, fieldDate, fieldValue, newStart, startTime, db, sdb)
	if err != nil {
		return 0, 0, err
	}
	return matched1, allOtherCount2, nil
}

func gridNewCustomers(ctx context.Context, accountID, teamID, entityName, fieldName, fieldDate string, fieldValue interface{}, duration string, db *sqlx.DB, sdb *database.SecDB) (int, error) {
	startTime, endTime := timeseries.Duration(duration)
	e, err := entity.RetrieveFixedEntity(ctx, db, accountID, teamID, entityName)
	if err != nil {
		return 0, err
	}
	fields := e.FieldsIgnoreError()

	conditionFields := make([]graphdb.Field, 0)
	for _, f := range fields {
		if f.Name == fieldName {
			conditionFields = append(conditionFields, *f.MakeGraphFieldPlain())
		}

		if f.Name == fieldDate {
			conditionFields = append(conditionFields, timeRange(f.Key, startTime, endTime))
		}
	}

	gSegment := graphdb.BuildGNode(accountID, e.ID, false).MakeBaseGNode("", conditionFields)
	var result *redisgraph.QueryResult
	result, err = graphdb.GetCount(sdb.GraphPool(), gSegment, false)
	if err != nil {
		return 0, err
	}
	cr := counts(result)
	return cr["total_count"], nil

}

func churn(ctx context.Context, accountID, teamID, entityName, fieldName, fieldDate string, fieldValue interface{}, startTime, endTime time.Time, db *sqlx.DB, sdb *database.SecDB) (int, int, error) {
	e, err := entity.RetrieveFixedEntity(ctx, db, accountID, teamID, entityName)
	if err != nil {
		return 0, 0, err
	}
	fields := e.FieldsIgnoreError()

	conditionFields := make([]graphdb.Field, 0)
	for _, f := range fields {
		if f.Name == fieldName {
			conditionFields = append(conditionFields, *f.MakeGraphFieldPlain())
		}

		if f.Name == fieldDate {
			conditionFields = append(conditionFields, timeRange(f.Key, startTime, endTime))
		}
	}

	gSegment := graphdb.BuildGNode(accountID, e.ID, false).MakeBaseGNode("", conditionFields)
	var result *redisgraph.QueryResult
	result, err = graphdb.GetCount(sdb.GraphPool(), gSegment, true)
	if err != nil {
		return 0, 0, err
	}
	cr := counts(result)

	matched, allOther := allButMatched(fieldValue, cr)
	return matched, allOther, nil
}

func delayedAccounts(ctx context.Context, accountID, teamID, entityName, dateFieldName, sourceEntityName, exp string, db *sqlx.DB, sdb *database.SecDB) (int, error) {
	se, err := entity.RetrieveFixedEntity(ctx, db, accountID, teamID, sourceEntityName)
	if err != nil {
		return 0, err
	}

	e, err := entity.RetrieveFixedEntity(ctx, db, accountID, teamID, entityName)
	if err != nil {
		return 0, err
	}
	fields := e.FieldsIgnoreError()

	conditionFields, err := makeConditionsFromExp(ctx, accountID, e.ID, exp, db, sdb)
	if err != nil {
		return 0, err
	}
	if se.ID != "" {
		conditionFields = append(conditionFields, sourceble(se.ID))
	}
	for _, f := range fields {
		if f.Name == dateFieldName {
			dateexp := fmt.Sprintf("{{%s.%s}} bf {%s}", e.ID, f.Key, "now")
			dateConditions, _ := makeConditionsFromExp(ctx, accountID, e.ID, dateexp, db, sdb)
			conditionFields = append(conditionFields, dateConditions...)
		}
	}

	gSegment := graphdb.BuildGNode(accountID, e.ID, false).MakeBaseGNode("", conditionFields)
	result, err := graphdb.GetFromParentCount(sdb.GraphPool(), gSegment)
	if err != nil {
		return 0, err
	}
	crAffected := counts(result)

	// gSegment = graphdb.BuildGNode(accountID, se.ID, false).MakeBaseGNode("", []graphdb.Field{})
	// result, err = graphdb.GetCount(sdb.GraphPool(), gSegment, false)
	// if err != nil {
	// 	return 0, err
	// }
	// crTotal := counts(result)

	return crAffected["total_count"], nil
}

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

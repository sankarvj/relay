package dbservice

import (
	"context"
	"fmt"
	"log"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	rg "github.com/redislabs/redisgraph-go"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
)

type SpiderService struct {
	pdb *sqlx.DB
	sdb *database.SecDB
}

func (spider SpiderService) Result(ctx context.Context, accountID, entityID string, sortby, direction string, page int, docount, useReturn bool, conditions []graphdb.Field) ([]item.Item, map[string]int, error) {
	segmentResult, countResult, err := spider.segment(ctx, accountID, entityID, sortby, direction, page, docount, useReturn, conditions)
	if err != nil {
		return nil, nil, err
	}
	items, err := itemsResp(ctx, spider.pdb, accountID, segmentResult)
	if err != nil {
		return nil, nil, err
	}
	var totalCount map[string]int
	if spider.CountEnabled(docount, page) {
		totalCount = counts(countResult)
	}

	return items, totalCount, nil
}

func (spider SpiderService) Count(ctx context.Context, accountID, entityID, groupById, groupLogic string, conditions []graphdb.Field) ([]Counters, error) {
	gSegment := graphdb.BuildGNode(accountID, entityID, false, nil).MakeBaseGNode("", conditions)

	var result *rg.QueryResult
	var err error
	switch groupLogic {
	case "g_b_id":
		result, err = graphdb.GetCount(spider.sdb.GraphPool(), gSegment, true)
	case "g_b_f":
		result, err = graphdb.GetGroupedCount(spider.sdb.GraphPool(), gSegment, groupById)
	case "g_b_f_r":
		result, err = graphdb.GetGroupedIDPlusFieldCount(spider.sdb.GraphPool(), gSegment, groupById, true)
	case "g_b_f_r2":
		result, err = graphdb.GetGroupedIDPlusFieldCount(spider.sdb.GraphPool(), gSegment, groupById, false)
	case "g_b_p":
		result, err = graphdb.GetFromParentCount(spider.sdb.GraphPool(), gSegment)
	default:
		result, err = graphdb.GetCount(spider.sdb.GraphPool(), gSegment, false)
	}

	if err != nil {
		return nil, err
	}
	counters := make([]Counters, 0)
	for result.Next() {
		r := result.Record()
		count := r.GetByIndex(0)
		itemID := "total_count"
		if r.GetByIndex(1) != nil {
			itemID = r.GetByIndex(1).(string)
		}
		var groupID string
		if r.GetByIndex(2) != nil {
			groupID = r.GetByIndex(2).(string)
		}
		counters = append(counters, Counters{ID: itemID, GroupID: groupID, Count: count})
	}
	return counters, nil
}

func (spider SpiderService) Sum(ctx context.Context, accountID, entityID, groupById string, conditions []graphdb.Field) ([]Counters, error) {
	//{Operator:in Key:uuid-00-contacts DataType:S Value:6eb4f58e-8327-4ccc-a262-22ad809e76cb}
	gSegment := graphdb.BuildGNode(accountID, entityID, false, nil).MakeBaseGNode("", conditions)
	result, err := graphdb.GetSum(spider.sdb.GraphPool(), gSegment, groupById)
	if err != nil {
		return nil, err
	}

	counters := make([]Counters, 0)
	for result.Next() {
		r := result.Record()
		count := r.GetByIndex(0)
		itemID := r.GetByIndex(1).(string)
		var groupID string
		if r.GetByIndex(2) != nil {
			groupID = r.GetByIndex(2).(string)
		}
		counters = append(counters, Counters{ID: itemID, GroupID: groupID, Count: count})
	}
	return counters, nil
}

func (rgs SpiderService) Search1(ctx context.Context, accountID, entityID string, conditionFields []graphdb.Field) []interface{} {
	result, _, err := rgs.segment(ctx, accountID, entityID, "", "", 0, false, true, conditionFields)
	if err != nil {
		return nil
	}
	return itemElements(result)
}

func (rgs SpiderService) Search2(ctx context.Context, accountID, entityID string, conditionFields []graphdb.Field) ([]item.Item, error) {
	result, _, err := rgs.segment(ctx, accountID, entityID, "", "", 0, false, false, conditionFields)
	if err != nil {
		return nil, nil
	}
	items, err := itemsResp(ctx, rgs.pdb, accountID, result)
	if err != nil {
		return nil, nil
	}
	return items, nil
}

func (rgs SpiderService) Search3(ctx context.Context, accountID, entityID string, conditionFields []graphdb.Field) []interface{} {
	result, _, err := rgs.segment(ctx, accountID, entityID, "", "", 0, false, false, conditionFields)
	if err != nil {
		return nil
	}
	itemIds := util.ParseGraphResult(result)
	return itemIds
}

func (rgs SpiderService) segment(ctx context.Context, accountID, entityID string, sortby, direction string, page int, docount, useReturn bool, conditions []graphdb.Field) (*rg.QueryResult, *rg.QueryResult, error) {
	//{Operator:in Key:uuid-00-contacts DataType:S Value:6eb4f58e-8327-4ccc-a262-22ad809e76cb}
	gSegment := graphdb.BuildGNode(accountID, entityID, false, nil).MakeBaseGNode("", conditions)
	gSegment.UseReturnNode = useReturn

	return listWithCountAsync(rgs.sdb.GraphPool(), gSegment, page, sortby, direction, rgs.CountEnabled(docount, page))
}

func listWithCountAsync(rp *redis.Pool, gSegment graphdb.GraphNode, page int, sortby, direction string, doCount bool) (*rg.QueryResult, *rg.QueryResult, error) {
	//going async way
	loopCount := 1
	type dbResult struct {
		result *rg.QueryResult
		_type  string
	}

	resc, errc := make(chan dbResult), make(chan error)
	go func(rPool *redis.Pool, gSegment graphdb.GraphNode, pageNo int, sortBy, direction string) {
		result, err := graphdb.GetResult(rp, gSegment, page, sortby, direction)
		if err != nil {
			errc <- err
			return
		}
		resc <- dbResult{result: result, _type: "segment"}
	}(rp, gSegment, page, sortby, direction)
	if doCount {
		loopCount = 2
		go func(rPool *redis.Pool, gCount graphdb.GraphNode) {
			result, err := graphdb.GetCount(rPool, gCount, false)
			if err != nil {
				errc <- err
				return
			}
			resc <- dbResult{result: result, _type: "count"}
		}(rp, gSegment)
	}

	var err error
	var segmentResult *rg.QueryResult
	var countResult *rg.QueryResult

	for i := 0; i < loopCount; i++ {
		select {
		case dbResult := <-resc:
			switch dbResult._type {
			case "segment":
				segmentResult = dbResult.result
			case "count":
				countResult = dbResult.result
			}
		case err := <-errc:
			fmt.Println(err)
		}
	}

	return segmentResult, countResult, err
}

func (rgs *SpiderService) CountEnabled(doCount bool, page int) bool {
	return doCount && page == 0
}

func itemsResp(ctx context.Context, db *sqlx.DB, accountID string, result *rg.QueryResult) ([]item.Item, error) {
	itemIDs := util.ParseGraphResult(result)
	items, err := item.BulkRetrieveItems(ctx, accountID, itemIDs, db)
	if err != nil {
		return []item.Item{}, err
	}

	return sort(items, itemIDs), nil
}

func sort(items []item.Item, itemIds []interface{}) []item.Item {
	itemMap := make(map[string]item.Item, len(items))
	for i := 0; i < len(items); i++ {
		itemMap[items[i].ID] = items[i]
	}
	sortedItems := make([]item.Item, 0)
	for _, id := range itemIds {
		sortedItems = append(sortedItems, itemMap[id.(string)])
	}
	return sortedItems
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

func publicRecordsOnly() graphdb.Field {
	return graphdb.Field{
		Key:        "system_is_public",
		Value:      true,
		DataType:   graphdb.TypeString,
		IsReverse:  false,
		Expression: "=",
	}
}

func itemElements(result *rg.QueryResult) []interface{} {
	values := make([]interface{}, 0)
	for result.Next() { // Next returns true until the iterator is depleted.
		// Get the current Record.
		r := result.Record()

		// Entries in the Record can be accessed by index or key.
		record := util.ConvertInterfaceToMap(util.ConvertInterfaceToMap(r.GetByIndex(0))["Properties"])
		log.Printf("record %+v", record)
		values = append(values, record["element"])
	}
	return values
}

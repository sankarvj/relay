package item

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

// TimeSeriesList retrieves a list of existing item for the entity associated from the database.
func TimeSeriesList(ctx context.Context, entityID string, db *sqlx.DB) ([]TimeSeriesItem, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.TimeSeriesList")
	defer span.End()

	items := []TimeSeriesItem{}
	const q = `SELECT fieldsb->>'9f9ade37-9549-4d12-a82d-c69495e85980' AS status, date_trunc('hour', ("fieldsb"->>'d3e572e1-3950-46db-a230-d41b2f4cd8d0')::timestamp) AS date , count(*) AS "value" from items where entity_id = $1 group by date,status;`

	if err := db.SelectContext(ctx, &items, q, entityID); err != nil {
		return nil, errors.Wrap(err, "selecting timeseries items")
	}

	return items, nil
}

//convert this to graphDB
func SearchByKey(ctx context.Context, entityID, key, term string, db *sqlx.DB) ([]Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.SearchByKey")
	defer span.End()

	items := []Item{}
	const q = `SELECT * FROM items where entity_id = $1`

	if err := db.SelectContext(ctx, &items, q, entityID); err != nil {
		return nil, errors.Wrap(err, "searching items")
	}

	return items, nil
}

func BulkRetrieve(ctx context.Context, entityID string, ids []interface{}, db *sqlx.DB) ([]Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.BulkRetrieve")
	defer span.End()

	items := []Item{}
	const q = `SELECT * FROM items where entity_id = $1 AND item_id = any($2)`

	if err := db.SelectContext(ctx, &items, q, entityID, pq.Array(ids)); err != nil {
		return items, errors.Wrap(err, "selecting bulk items for entity id and selected item ids")
	}

	return items, nil
}

//TimeSeriesSameDayViewModel presents the item inside a time ticker map
func TimeSeriesSameDayViewModel(items []TimeSeriesItem, start time.Time, loop int) map[time.Time]TimeSeriesItem {
	timeSeriesMap := make(map[time.Time]TimeSeriesItem, 0)
	for _, item := range items {
		timeSeriesMap[item.Date] = item
	}
	rounded := time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), 0, 0, 0, start.Location())
	for i := 0; i < loop; i++ {
		if _, ok := timeSeriesMap[rounded]; !ok {
			timeSeriesMap[rounded] = TimeSeriesItem{
				State: "down",
				Date:  rounded,
				Value: 0,
			}
		}
		rounded = rounded.Add(time.Hour * 1)
	}
	return timeSeriesMap
}

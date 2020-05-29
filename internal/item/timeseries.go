package item

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

// TimeSeriesList retrieves a list of existing item for the entity associated from the database.
func TimeSeriesList(ctx context.Context, entityID string, db *sqlx.DB) ([]TimeSeriesItem, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.TimeSeriesList")
	defer span.End()

	items := []TimeSeriesItem{}
	const q = `SELECT input->>'uuuid3' AS status, date_trunc('hour', ("input"->>'uuuid1')::timestamp) AS date , count(*) AS "value" from items where entity_id = $1 group by date,status;`

	if err := db.SelectContext(ctx, &items, q, entityID); err != nil {
		return nil, errors.Wrap(err, "selecting timeseries items")
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
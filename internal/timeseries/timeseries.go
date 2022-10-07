package timeseries

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

var (
	ErrTimeseriesNotFound = errors.New("Timeseries ID not found")
	// ErrInvalidTimeSeriesID occurs when an ID is not in a valid form.
	ErrInvalidTimeSeriesID = errors.New("Timeseries ID is not in its proper form")
)

func Create(ctx context.Context, db *sqlx.DB, nt NewTimeseries, now time.Time) (Timeseries, error) {
	ctx, span := trace.StartSpan(ctx, "internal.timeseries.Create")
	defer span.End()

	fieldsBytes, err := json.Marshal(nt.Fields)
	if err != nil {
		return Timeseries{}, errors.Wrap(err, "encode fields to bytes")
	}

	if nt.Identifier != nil && *nt.Identifier == "" {
		nt.Identifier = nil
	}

	t := Timeseries{
		ID:          nt.ID,
		AccountID:   nt.AccountID,
		EntityID:    nt.EntityID,
		Type:        nt.Type,
		Identifier:  nt.Identifier,
		Tags:        nt.Tags,
		Event:       nt.Event,
		Description: nt.Description,
		Count:       nt.Count,
		StartTime:   now.UTC(),
		EndTime:     now.UTC(),
		Fieldsb:     string(fieldsBytes),
	}

	const q = `INSERT INTO timeseries
		(timeseries_id, account_id, entity_id, type, identifier, tags, event, description, count, start_time, end_time, fieldsb)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err = db.ExecContext(
		ctx, q,
		t.ID, t.AccountID, t.EntityID, t.Type, t.Identifier, t.Tags, t.Event, t.Description, t.Count, t.StartTime, t.EndTime, t.Fieldsb,
	)
	if err != nil {
		return Timeseries{}, errors.Wrap(err, "inserting timeseries data")
	}

	return t, nil
}

func Update(ctx context.Context, db *sqlx.DB, ts Timeseries, now time.Time) (*Timeseries, error) {
	ctx, span := trace.StartSpan(ctx, "internal.entity.Update")
	defer span.End()

	const q = `UPDATE timeseries SET
		"fieldsb" = $4,
		"count" = $5,
		"end_time" = $6 
		WHERE account_id = $1 AND entity_id = $2 AND timeseries_id =$3`
	_, err := db.ExecContext(ctx, q, ts.AccountID, ts.EntityID, ts.ID, ts.Fieldsb, ts.Count, now.UTC())
	if err != nil {
		return nil, err
	}

	return &ts, nil
}

func List(ctx context.Context, accountID, entityID string, startTime, endTime time.Time, db *sqlx.DB) ([]Timeseries, error) {
	ctx, span := trace.StartSpan(ctx, "internal.timeseries.List")
	defer span.End()

	timeseries := []Timeseries{}
	const q = `SELECT * FROM timeseries where account_id = $1 AND entity_id = $2 AND start_time > $3 AND end_time < $4`

	if err := db.SelectContext(ctx, &timeseries, q, accountID, entityID, startTime, endTime); err != nil {
		return nil, errors.Wrap(err, "selecting timeseries")
	}

	return timeseries, nil
}

func Count(ctx context.Context, accountID, entityID string, startTime, endTime time.Time, db *sqlx.DB) (int, error) {
	ctx, span := trace.StartSpan(ctx, "internal.timeseries.Count")
	defer span.End()

	var count int
	const q = `SELECT count(*) FROM timeseries where account_id = $1 AND entity_id = $2 AND start_time > $3 AND end_time < $4`
	if err := db.GetContext(ctx, &count, q, accountID, entityID, startTime, endTime); err != nil {
		if err == sql.ErrNoRows {
			return count, nil
		}
		return count, errors.Wrapf(err, "selecting count")
	}
	return count, nil
}

func RetriveLatest(ctx context.Context, accountID, entityID string, identifier *string, db *sqlx.DB) ([]Timeseries, error) {
	ctx, span := trace.StartSpan(ctx, "internal.timeseries.List")
	defer span.End()

	timeseries := []Timeseries{}

	if identifier == nil || *identifier == "" {
		const q = `SELECT * FROM timeseries where account_id = $1 AND entity_id = $2 ORDER BY end_time DESC LIMIT 1`
		if err := db.SelectContext(ctx, &timeseries, q, accountID, entityID); err != nil {
			return nil, errors.Wrap(err, "retriveing latest timeseries")
		}
	} else {
		const q = `SELECT * FROM timeseries where account_id = $1 AND entity_id = $2 AND identifier = $3 ORDER BY end_time DESC LIMIT 1`
		if err := db.SelectContext(ctx, &timeseries, q, accountID, entityID, identifier); err != nil {
			return nil, errors.Wrap(err, "retriveing latest timeseries")
		}
	}

	return timeseries, nil
}

func Retrieve(ctx context.Context, accountID, entityID, timeseriesID string, db *sqlx.DB) (Timeseries, error) {
	ctx, span := trace.StartSpan(ctx, "internal.timeseries.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(timeseriesID); err != nil {
		return Timeseries{}, ErrInvalidTimeSeriesID
	}

	var t Timeseries
	const q = `SELECT * FROM timeseries WHERE account_id = $1 AND entity_id = $2 AND timeseries_id = $3`
	if err := db.GetContext(ctx, &t, q, accountID, entityID, timeseriesID); err != nil {
		if err == sql.ErrNoRows {
			return Timeseries{}, ErrTimeseriesNotFound
		}

		return Timeseries{}, errors.Wrapf(err, "selecting timeseries %q", timeseriesID)
	}

	return t, nil
}

// Fields parses attribures to fields
func (t Timeseries) Fields() map[string]interface{} {
	var fields map[string]interface{}
	if t.Fieldsb == "" {
		return make(map[string]interface{}, 0)
	}
	if err := json.Unmarshal([]byte(t.Fieldsb), &fields); err != nil {
		log.Printf("***> unexpected error occurred when unmarshalling fields for timeseries: %v error: %v\n", t.ID, err)
		return make(map[string]interface{}, 0)
	}
	return fields
}

func Duration(duration string) (time.Time, time.Time, time.Time) {
	loc, _ := time.LoadLocation("UTC")
	return DurationWithZone(duration, loc)
}

func DurationWithZone(duration string, zone *time.Location) (time.Time, time.Time, time.Time) {
	t := time.Now().In(zone)
	endTime := t
	stTime := endTime
	switch duration {
	case "last_hr":
		stTime = endTime.Add(-time.Hour * 1)
	case "last_6hr":
		stTime = endTime.Add(-time.Hour * 6)
	case "last_24hrs":
		stTime = endTime.Add(-time.Hour * 24)
	case "today":
		stTime = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	case "yesterday":
		stTime = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
		stTime = stTime.Add(-time.Hour * 24)
	case "this_week":
		stTime = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
		stTime = endTime.AddDate(0, 0, int(t.Weekday()))
	case "last_week":
		stTime = endTime.AddDate(0, 0, -7)
	case "this_month":
		stTime = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
		stTime = endTime.AddDate(0, 0, -t.Day())
	case "last_month":
		stTime = endTime.AddDate(0, 0, -30)
	case "last_6month":
		stTime = endTime.AddDate(0, 0, -1)
	}
	difference := stTime.Sub(endTime)
	lastStart := stTime.Add(difference)
	return stTime.UTC(), endTime.UTC(), lastStart.UTC()
}

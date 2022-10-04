package chart

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

var (
	// ErrChartNotFound is used when a specific chart is requested but does not exist.
	ErrChartNotFound = errors.New("Chart not found")
)

func List(ctx context.Context, accountID, group string, db *sqlx.DB) ([]Chart, error) {
	ctx, span := trace.StartSpan(ctx, "internal.chart.List")
	defer span.End()

	charts := []Chart{}
	if group == "" {
		const q = `SELECT * FROM charts where account_id = $1 LIMIT 50`
		if err := db.SelectContext(ctx, &charts, q, accountID); err != nil {
			return charts, errors.Wrap(err, "selecting charts for an account")
		}
	} else {
		const q = `SELECT * FROM charts where account_id = $1 AND group = $2 LIMIT 50`
		if err := db.SelectContext(ctx, &charts, q, accountID, group); err != nil {
			return charts, errors.Wrap(err, "selecting charts for an account with group")
		}
	}

	return charts, nil
}

func Create(ctx context.Context, db *sqlx.DB, nc NewChart, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.chart.Create")
	defer span.End()
	t := Chart{
		ID:             uuid.New().String(),
		AccountID:      nc.AccountID,
		EntityID:       nc.EntityID,
		ParentEntityID: nc.ParentEntityID,
		UserID:         nc.UserID,
		Field:          nc.Field,
		Name:           nc.Name,
		Type:           nc.Type,
		Group:          nc.Group,
		Duration:       nc.Duration,
		State:          nc.State,
		Calc:           nc.Calc,
		Position:       nc.Position,
		CreatedAt:      now.UTC(),
	}

	const q = `INSERT INTO charts
		(chart_id, account_id, entity_id, parent_entity_id, user_id, field, name, type, group, duration, state, calc, position, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
	_, err := db.ExecContext(
		ctx, q,
		t.ID, t.AccountID, t.EntityID, t.ParentEntityID, t.UserID, t.Field, t.Name, t.Type, t.Group, t.Duration, t.State, t.Calc, t.Position,
		t.CreatedAt,
	)
	if err != nil {
		return errors.Wrap(err, "chart created")
	}

	return nil
}

func Delete(ctx context.Context, db *sqlx.DB, chartID string) error {
	ctx, span := trace.StartSpan(ctx, "internal.chart.Delete")
	defer span.End()

	const q = `DELETE FROM charts WHERE chart_id = $1`

	if _, err := db.ExecContext(ctx, q, chartID); err != nil {
		return errors.Wrapf(err, "chart delete")
	}

	return nil
}

func Retrieve(ctx context.Context, db *sqlx.DB, accountID, chartID string) (*Chart, error) {
	ctx, span := trace.StartSpan(ctx, "internal.chart.Retrieve")
	defer span.End()

	var c Chart
	const q = `SELECT * FROM charts WHERE account_id =$1 AND chart_id = $2`
	if err := db.GetContext(ctx, &c, q, accountID, chartID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrChartNotFound
		}

		return nil, err
	}

	return &c, nil
}

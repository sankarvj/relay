package rule

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

// List retrieves a list of existing rules for the entity associated from the database.
func List(ctx context.Context, entityID string, db *sqlx.DB) ([]Rule, error) {
	ctx, span := trace.StartSpan(ctx, "internal.rule.List")
	defer span.End()

	rules := []Rule{}
	const q = `SELECT * FROM rules where entity_id = $1`

	if err := db.SelectContext(ctx, &rules, q, entityID); err != nil {
		return nil, errors.Wrap(err, "selecting rules")
	}

	return rules, nil
}

// Create inserts a new item into the database.
func Create(ctx context.Context, db *sqlx.DB, n NewRule, now time.Time) (*Rule, error) {
	ctx, span := trace.StartSpan(ctx, "internal.rule.Create")
	defer span.End()

	// actionItems, err := json.Marshal(n.ActionItems)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "encode action")
	// }

	r := Rule{
		ID:       uuid.New().String(),
		EntityID: n.EntityID,
		//Expression: n.Expression + " <" + string(actionItems) + ">",
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC().Unix(),
	}

	const q = `INSERT INTO rules
		(rule_id, entity_id, expression, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := db.ExecContext(
		ctx, q,
		r.ID, r.EntityID, r.Expression,
		r.CreatedAt, r.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "inserting rule")
	}

	return &r, nil
}

package item

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

// List retrieves a list of existing item for the entity associated from the database.
func List(ctx context.Context, entityID string, db *sqlx.DB) ([]Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.List")
	defer span.End()

	items := []Item{}
	const q = `SELECT * FROM items where entity_id = $1`

	if err := db.SelectContext(ctx, &items, q, entityID); err != nil {
		return nil, errors.Wrap(err, "selecting items")
	}

	return items, nil
}

// Create inserts a new item into the database.
func Create(ctx context.Context, db *sqlx.DB, n NewItem, now time.Time) (*Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.Create")
	defer span.End()

	input, err := json.Marshal(n.Fields)
	if err != nil {
		return nil, errors.Wrap(err, "encode fields to input")
	}

	i := Item{
		ID:        uuid.New().String(),
		EntityID:  n.EntityID,
		Input:     string(input),
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC().Unix(),
	}

	const q = `INSERT INTO items
		(item_id, entity_id, input, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err = db.ExecContext(
		ctx, q,
		i.ID, i.EntityID, i.Input,
		i.CreatedAt, i.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "inserting item")
	}

	return &i, nil
}

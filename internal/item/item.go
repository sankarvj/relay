package item

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

var (
	// ErrNotFound is used when a specific entity is requested but does not exist.
	ErrNotFound = errors.New("Item not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")
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
func Create(ctx context.Context, db *sqlx.DB, entityID string, n NewItem, now time.Time) (*Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.Create")
	defer span.End()

	input, err := json.Marshal(n.Fields)
	if err != nil {
		return nil, errors.Wrap(err, "encode fields to input")
	}

	i := Item{
		ID:        uuid.New().String(),
		EntityID:  entityID,
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

//UpdateFields patches the field data
func UpdateFields(ctx context.Context, db *sqlx.DB, id string, fields map[string]interface{}) error {
	input, err := json.Marshal(fields)
	if err != nil {
		return errors.Wrap(err, "encode fields to input")
	}
	inputStr := string(input)
	upd := UpdateItem{
		Input: &inputStr,
	}
	return update(ctx, db, id, upd, time.Now())
}

// Update replaces a item document in the database.
func update(ctx context.Context, db *sqlx.DB, id string, upd UpdateItem, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.item.Update")
	defer span.End()

	i, err := Retrieve(ctx, id, db)
	if err != nil {
		return err
	}

	if upd.Input != nil {
		i.Input = *upd.Input
	}
	i.UpdatedAt = now.Unix()

	const q = `UPDATE items SET
		"input" = $2,
		"updated_at" = $3
		WHERE item_id = $1`
	_, err = db.ExecContext(ctx, q, i.ID,
		i.Input, i.UpdatedAt,
	)
	if err != nil {
		return errors.Wrap(err, "updating item")
	}

	return nil
}

// Retrieve gets the specified user from the database.
func Retrieve(ctx context.Context, id string, db *sqlx.DB) (*Item, error) {
	ctx, span := trace.StartSpan(ctx, "internal.item.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidID
	}

	var i Item
	const q = `SELECT * FROM items WHERE item_id = $1`
	if err := db.GetContext(ctx, &i, q, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, errors.Wrapf(err, "selecting item %q", id)
	}

	return &i, nil
}

package entity

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"strconv"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

var (
	// ErrNotFound is used when a specific entity is requested but does not exist.
	ErrNotFound = errors.New("Entity not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")
)

// List retrieves a list of existing entities for the team associated from the database.
func List(ctx context.Context, teamID string, db *sqlx.DB) ([]Entity, error) {
	ctx, span := trace.StartSpan(ctx, "internal.entity.List")
	defer span.End()

	entities := []Entity{}
	const q = `SELECT * FROM entities where team_id = $1`

	if err := db.SelectContext(ctx, &entities, q, teamID); err != nil {
		return nil, errors.Wrap(err, "selecting entities")
	}

	return entities, nil
}

// Primary retrieves the primary entity for the team associated from the database.
func Primary(ctx context.Context, teamID string, db *sqlx.DB) (Entity, error) {
	ctx, span := trace.StartSpan(ctx, "internal.entity.Primary")
	defer span.End()

	entities := []Entity{}
	const q = `SELECT * FROM entities where team_id = $1 and mode = $2 limit 1`
	if err := db.SelectContext(ctx, &entities, q, teamID, ModePrimary); err != nil {
		return Entity{}, errors.Wrap(err, "selecting entities")
	}

	if len(entities) > 0 {
		return entities[0], nil
	}

	return Entity{}, errors.New("no primary entity present")
}

// Create inserts a new user into the database.
func Create(ctx context.Context, db *sqlx.DB, n NewEntity, now time.Time) (*Entity, error) {
	ctx, span := trace.StartSpan(ctx, "internal.entity.Create")
	defer span.End()

	teamID, _ := strconv.ParseInt(n.TeamID, 10, 64)

	attributes, err := json.Marshal(n.Fields)
	if err != nil {
		return nil, errors.Wrap(err, "encode fields to attributes")
	}

	e := Entity{
		ID:         uuid.New().String(),
		TeamID:     teamID,
		Name:       n.Name,
		Attributes: string(attributes),
		CreatedAt:  now.UTC(),
		UpdatedAt:  now.UTC().Unix(),
	}

	const q = `INSERT INTO entities
		(entity_id, team_id, name, attributes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err = db.ExecContext(
		ctx, q,
		e.ID, e.TeamID, e.Name, e.Attributes,
		e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "inserting entity")
	}

	return &e, nil
}

// Retrieve gets the specified entity from the database.
func Retrieve(ctx context.Context, id string, db *sqlx.DB) (*Entity, error) {
	ctx, span := trace.StartSpan(ctx, "internal.entity.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidID
	}

	var e Entity
	const q = `SELECT * FROM entities WHERE entity_id = $1`
	if err := db.GetContext(ctx, &e, q, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, errors.Wrapf(err, "selecting entity %q", id)
	}

	return &e, nil
}

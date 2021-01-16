package entity

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/relationship"
	"go.opencensus.io/trace"
)

var (
	// ErrNotFound is used when a specific entity is requested but does not exist.
	ErrNotFound = errors.New("Entity not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")
)

// List retrieves a list of existing entities for the team associated from the database.
func List(ctx context.Context, teamID string, categoryIds []int, db *sqlx.DB) ([]Entity, error) {
	ctx, span := trace.StartSpan(ctx, "internal.entity.List")
	defer span.End()
	entities := []Entity{}
	if len(categoryIds) == 0 {
		const q = `SELECT * FROM entities where team_id = $1`
		if err := db.SelectContext(ctx, &entities, q, teamID); err != nil {
			return nil, errors.Wrap(err, "selecting entities for all category")
		}
	} else {
		const q = `SELECT * FROM entities where team_id = $1 AND category = any($2)`
		if err := db.SelectContext(ctx, &entities, q, teamID, pq.Array(categoryIds)); err != nil {
			return nil, errors.Wrap(err, "selecting entities for category")
		}
	}

	return entities, nil
}

// Create inserts a new user into the database.
func Create(ctx context.Context, db *sqlx.DB, n NewEntity, now time.Time) (Entity, error) {
	ctx, span := trace.StartSpan(ctx, "internal.entity.Create")
	defer span.End()

	fieldsBytes, err := json.Marshal(n.Fields)
	if err != nil {
		return Entity{}, errors.Wrap(err, "encode fields to bytes")
	}

	e := Entity{
		ID:          n.ID,
		AccountID:   n.AccountID,
		TeamID:      n.TeamID,
		Name:        n.Name,
		DisplayName: n.DisplayName,
		Category:    n.Category,
		State:       n.State,
		Fieldsb:     string(fieldsBytes),
		CreatedAt:   now.UTC(),
		UpdatedAt:   now.UTC().Unix(),
	}

	const q = `INSERT INTO entities
		(entity_id, account_id, team_id, name, display_name, category, state, fieldsb, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err = db.ExecContext(
		ctx, q,
		e.ID, e.AccountID, e.TeamID, e.Name, e.DisplayName, e.Category, e.State, e.Fieldsb,
		e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return Entity{}, errors.Wrap(err, "inserting entity")
	}

	//TODO: do it in the same transaction.
	//TODO: this relationship should happen only if the user explicitly specifies that.
	//may be, we can give add the boolean in the meta to identify that.
	err = relationship.Bonding(ctx, db, e.AccountID, e.ID, refFields(n.Fields))

	return e, err
}

// Update replaces a item document in the database.
func Update(ctx context.Context, db *sqlx.DB, entityID string, fieldsB string, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.entity.Update")
	defer span.End()

	e, err := Retrieve(ctx, entityID, db)
	if err != nil {
		return err
	}

	e.UpdatedAt = now.Unix()
	e.Fieldsb = fieldsB

	const q = `UPDATE entities SET
		"fieldsb" = $2,
		"updated_at" = $3
		WHERE entity_id = $1`
	_, err = db.ExecContext(ctx, q, e.ID,
		e.Fieldsb, e.UpdatedAt,
	)
	if err != nil {
		return err
	}

	//TODO: do it in the same transaction.
	updatedFields, err := e.AllFields()
	if err != nil {
		return err
	}
	err = relationship.Bonding(ctx, db, e.AccountID, e.ID, refFields(updatedFields))

	return nil
}

// Retrieve gets the specified entity from the database.
func Retrieve(ctx context.Context, id string, db *sqlx.DB) (Entity, error) {
	ctx, span := trace.StartSpan(ctx, "internal.entity.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return Entity{}, ErrInvalidID
	}

	var e Entity
	const q = `SELECT * FROM entities WHERE entity_id = $1`
	if err := db.GetContext(ctx, &e, q, id); err != nil {
		if err == sql.ErrNoRows {
			return Entity{}, ErrNotFound
		}

		return Entity{}, errors.Wrapf(err, "selecting entity %q", id)
	}

	return e, nil
}

func BulkRetrieve(ctx context.Context, ids []string, db *sqlx.DB) ([]Entity, error) {
	ctx, span := trace.StartSpan(ctx, "internal.entity.BulkRetrieve")
	defer span.End()

	entities := []Entity{}
	const q = `SELECT * FROM entities where entity_id = any($1)`

	if err := db.SelectContext(ctx, &entities, q, pq.Array(ids)); err != nil {
		return entities, errors.Wrap(err, "selecting bulk entities")
	}

	return entities, nil
}

func FetchIDs(entities []Entity) []string {
	ids := make([]string, 0)
	for _, e := range entities {
		ids = append(ids, e.ID)
	}
	return ids
}

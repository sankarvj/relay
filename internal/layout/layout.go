package layout

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
	ErrNotFound = errors.New("Layout not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("Layout ID is not in its proper form")
)

func Create(ctx context.Context, db *sqlx.DB, nl NewLayout, now time.Time) (Layout, error) {
	ctx, span := trace.StartSpan(ctx, "internal.layout.Create")
	defer span.End()

	fieldsBytes, err := json.Marshal(nl.Fields)
	if err != nil {
		return Layout{}, errors.Wrap(err, "encode fields to bytes")
	}

	l := Layout{
		Name:      nl.Name,
		AccountID: nl.AccountID,
		EntityID:  nl.EntityID,
		UserID:    nl.UserID,
		Type:      nl.Type,
		Fieldsb:   string(fieldsBytes),
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC().Unix(),
	}

	const q = `INSERT INTO layouts
		(name, account_id, entity_id, user_id, type, fieldsb, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = db.ExecContext(
		ctx, q,
		l.Name, l.AccountID, l.EntityID, l.UserID, l.Type, l.Fieldsb,
		l.CreatedAt, l.UpdatedAt,
	)
	if err != nil {
		return Layout{}, errors.Wrap(err, "inserting layout")
	}

	return l, nil
}

func Retrieve(ctx context.Context, accountID, entityID, name string, db *sqlx.DB) (Layout, error) {
	ctx, span := trace.StartSpan(ctx, "internal.layout.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(name); err != nil {
		return Layout{}, ErrInvalidID
	}

	var l Layout
	const q = `SELECT * FROM layouts WHERE account_id = $1 AND entity_id = $2 AND name = $3`
	if err := db.GetContext(ctx, &l, q, accountID, entityID, name); err != nil {
		if err == sql.ErrNoRows {
			return Layout{}, ErrNotFound
		}

		return Layout{}, errors.Wrapf(err, "selecting layout %q", name)
	}

	return l, nil
}

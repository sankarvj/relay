package visitor

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
	// ErrNotFound is used when a specific entity is requested but does not exist.
	ErrNotFound = errors.New("Visitor not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("Visitor ID is not in its proper form")
)

func List(ctx context.Context, accountID, email string, db *sqlx.DB) ([]Visitor, error) {
	ctx, span := trace.StartSpan(ctx, "internal.visitor.List")
	defer span.End()
	visitors := []Visitor{}

	const q = `SELECT * FROM visitors where account_id = $1 AND email = $2`
	if err := db.SelectContext(ctx, &visitors, q, accountID, email); err != nil {
		return nil, errors.Wrap(err, "selecting visitors for an user")
	}
	return visitors, nil
}

func Create(ctx context.Context, db *sqlx.DB, nv NewVisitor, now time.Time) (Visitor, error) {
	ctx, span := trace.StartSpan(ctx, "internal.visitor.Create")
	defer span.End()

	v := Visitor{
		VistitorID: uuid.New().String(),
		AccountID:  nv.AccountID,
		TeamID:     nv.TeamID,
		EntityID:   nv.EntityID,
		ItemID:     nv.ItemID,
		Email:      nv.Email,
		Token:      nv.Token,
		CreatedAt:  now.UTC(),
		UpdatedAt:  now.UTC().Unix(),
	}

	const q = `INSERT INTO visitors
		(visitor_id, account_id, team_id, entity_id, item_id, email, token, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := db.ExecContext(
		ctx, q,
		v.VistitorID, v.AccountID, v.TeamID, v.EntityID, v.ItemID, v.Email, v.Token,
		v.CreatedAt, v.UpdatedAt,
	)
	if err != nil {
		return Visitor{}, errors.Wrap(err, "inserting visitor")
	}

	return v, nil
}

func Retrieve(ctx context.Context, accountID, visitorID string, db *sqlx.DB) (Visitor, error) {
	ctx, span := trace.StartSpan(ctx, "internal.layout.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(visitorID); err != nil {
		return Visitor{}, ErrInvalidID
	}

	var v Visitor
	const q = `SELECT * FROM visitors WHERE account_id = $1 AND visitor_id = $2`
	if err := db.GetContext(ctx, &v, q, accountID, visitorID); err != nil {
		if err == sql.ErrNoRows {
			return Visitor{}, ErrNotFound
		}

		return Visitor{}, errors.Wrapf(err, "selecting visitor %q", visitorID)
	}

	return v, nil
}

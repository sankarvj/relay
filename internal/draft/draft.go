package draft

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
	// ErrNotFound is used when a specific User is requested but does not exist.
	ErrNotFound = errors.New("Draft not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")
)

// Retrieve gets the specified account from the database.
func Retrieve(ctx context.Context, draftID string, db *sqlx.DB) (*Draft, error) {
	ctx, span := trace.StartSpan(ctx, "internal.draft.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(draftID); err != nil {
		return nil, ErrInvalidID
	}

	var d Draft
	const q = `SELECT * FROM drafts WHERE draft_id = $1`
	if err := db.GetContext(ctx, &d, q, draftID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, errors.Wrapf(err, "selecting draft %q", draftID)
	}

	return &d, nil
}

// Create inserts a new user into the database. Call AccountBootstrap instead
func Create(ctx context.Context, nd NewDraft, now time.Time, db *sqlx.DB) (Draft, error) {
	d := Draft{
		ID:            uuid.New().String(),
		AccountName:   nd.AccountName,
		BusinessEmail: nd.BusinessEmail,
		Teams:         nd.Teams,
		CreatedAt:     now.UTC(),
		UpdatedAt:     now.UTC().Unix(),
	}

	const q = `INSERT INTO drafts
		(draft_id, account_name, business_email, teams, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.ExecContext(
		ctx, q,
		d.ID, d.AccountName, d.BusinessEmail, d.Teams,
		d.CreatedAt, d.UpdatedAt,
	)
	if err != nil {
		return Draft{}, errors.Wrap(err, "inserting draft")
	}

	return d, nil
}

func Delete(ctx context.Context, draftID string, db *sqlx.DB) error {
	ctx, span := trace.StartSpan(ctx, "internal.draft.Delete")
	defer span.End()

	const q = `DELETE FROM drafts WHERE draft_id = $1`

	if _, err := db.ExecContext(ctx, q, draftID); err != nil {
		return errors.Wrapf(err, "deleting draft %s", draftID)
	}

	return nil
}

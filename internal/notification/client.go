package notification

import (
	"context"
	"database/sql"
	"time"

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

func CreateClient(ctx context.Context, db *sqlx.DB, cr ClientRegister, now time.Time) (ClientRegister, error) {
	ctx, span := trace.StartSpan(ctx, "internal.client.Create")
	defer span.End()

	cr.CreatedAt = now.UTC()
	cr.UpdatedAt = now.UTC().Unix()

	const q = `INSERT INTO clients
		(account_id, user_id, device_token, device_type, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := db.ExecContext(
		ctx, q,
		cr.AccountID, cr.UserID, cr.DeviceToken, cr.DeviceType, cr.Status,
		cr.CreatedAt, cr.UpdatedAt,
	)
	if err != nil {
		return ClientRegister{}, errors.Wrap(err, "inserting client")
	}

	return cr, nil
}

func RetrieveClient(ctx context.Context, accountID, userID string, db *sqlx.DB) (ClientRegister, error) {
	ctx, span := trace.StartSpan(ctx, "internal.client.Retrieve")
	defer span.End()

	var cr ClientRegister
	const q = `SELECT * FROM clients WHERE account_id = $1 AND user_id = $2`
	if err := db.GetContext(ctx, &cr, q, accountID, userID); err != nil {
		if err == sql.ErrNoRows {
			return ClientRegister{}, ErrNotFound
		}

		return ClientRegister{}, errors.Wrapf(err, "selecting client for userID %q", userID)
	}

	return cr, nil
}

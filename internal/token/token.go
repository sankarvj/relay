package token

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

var (
	// ErrTokenNotFound is used when a specific token is requested but does not exist.
	ErrTokenNotFound = errors.New("Token not found")
)

func Create(ctx context.Context, db *sqlx.DB, token, accountID string, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.token.Create")
	defer span.End()
	t := Token{
		Token:     token,
		AccountID: accountID,
		Type:      0,
		State:     0,
		Scope:     []string{},
		IssuedAt:  now.UTC(),
		Expiry:    now.UTC().Add(time.Hour * 24 * 7 * time.Duration(1000)), //roughly 20 years
		CreatedAt: now.UTC(),
	}

	const q = `INSERT INTO tokens
		(token, account_id, type, state, scope, issued_at, expiry, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := db.ExecContext(
		ctx, q,
		t.Token, t.AccountID, t.Type,
		t.State, t.Scope,
		t.IssuedAt, t.Expiry, t.CreatedAt,
	)
	if err != nil {
		return errors.Wrap(err, "token generation")
	}

	return nil
}

func Delete(ctx context.Context, db *sqlx.DB, token string) error {
	ctx, span := trace.StartSpan(ctx, "internal.token.Delete")
	defer span.End()

	const q = `DELETE FROM tokens WHERE token = $1`

	if _, err := db.ExecContext(ctx, q, token); err != nil {
		return errors.Wrapf(err, "token delete")
	}

	return nil
}

func Retrieve(ctx context.Context, db *sqlx.DB, accountID string) (*Token, error) {
	ctx, span := trace.StartSpan(ctx, "internal.token.Retrieve")
	defer span.End()

	var t Token
	const q = `SELECT * FROM tokens WHERE account_id = $1`
	if err := db.GetContext(ctx, &t, q, accountID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTokenNotFound
		}

		return nil, err
	}

	return &t, nil
}

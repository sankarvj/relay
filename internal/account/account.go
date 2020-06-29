package account

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

// List retrieves a list of existing accounts from the database.
func List(ctx context.Context, currentUserID string, db *sqlx.DB) ([]Account, error) {
	ctx, span := trace.StartSpan(ctx, "internal.account.List")
	defer span.End()

	accounts := []Account{}
	const q = `SELECT a.* FROM accounts as a join users as u on a.account_id = u.account_id where u.user_id = $1`

	if err := db.SelectContext(ctx, &accounts, q, currentUserID); err != nil {
		return nil, errors.Wrap(err, "selecting accounts")
	}
	return accounts, nil
}

// Create inserts a new user into the database.
func Create(ctx context.Context, db *sqlx.DB, n NewAccount, now time.Time) (Account, error) {
	ctx, span := trace.StartSpan(ctx, "internal.account.Create")
	defer span.End()

	a := Account{
		ID:        uuid.New().String(),
		Domain:    n.Domain,
		Name:      n.Name,
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC().Unix(),
	}

	const q = `INSERT INTO accounts
		(account_id, name, domain, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := db.ExecContext(
		ctx, q,
		a.ID, a.Name, a.Domain,
		a.CreatedAt, a.UpdatedAt,
	)
	if err != nil {
		return Account{}, errors.Wrap(err, "inserting account")
	}

	return a, nil
}

package account

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
	ErrNotFound = errors.New("Account not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")
)

// List retrieves a list of existing accounts from the database.
func List(ctx context.Context, currentUserID string, db *sqlx.DB) ([]Account, error) {
	ctx, span := trace.StartSpan(ctx, "internal.account.List")
	defer span.End()

	accounts := []Account{}
	const q = `SELECT a.* FROM accounts as a join users as u on a.account_id = ANY (u.account_ids) where u.user_id = $1`

	if err := db.SelectContext(ctx, &accounts, q, currentUserID); err != nil {
		return nil, errors.Wrap(err, "selecting accounts")
	}
	return accounts, nil
}

// Retrieve gets the specified account from the database.
func Retrieve(ctx context.Context, db *sqlx.DB, id string) (*Account, error) {
	ctx, span := trace.StartSpan(ctx, "internal.account.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidID
	}

	var a Account
	const q = `SELECT * FROM accounts WHERE account_id = $1`
	if err := db.GetContext(ctx, &a, q, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, errors.Wrapf(err, "selecting account %q", id)
	}

	return &a, nil
}

func CheckAvailability(ctx context.Context, name string, db *sqlx.DB) (*Account, error) {
	ctx, span := trace.StartSpan(ctx, "internal.account.CheckAvailability")
	defer span.End()

	var a Account
	const q = `SELECT * FROM accounts WHERE LOWER(name) = LOWER($1)`
	if err := db.GetContext(ctx, &a, q, name); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, errors.Wrapf(err, "checking account availability %s", name)
	}

	return &a, nil
}

// Create inserts a new user into the database. Call AccountBootstrap instead
func Create(ctx context.Context, db *sqlx.DB, n NewAccount, now time.Time) (Account, error) {
	a := Account{
		ID:        n.ID,
		Name:      n.Name,
		Domain:    n.Domain,
		IssuedAt:  now.UTC(),
		Expiry:    now.UTC(),
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC().Unix(),
	}

	const q = `INSERT INTO accounts
		(account_id, name, domain, issued_at, expiry, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := db.ExecContext(
		ctx, q,
		a.ID, a.Name, a.Domain,
		a.IssuedAt, a.Expiry,
		a.CreatedAt, a.UpdatedAt,
	)
	if err != nil {
		return Account{}, errors.Wrap(err, "inserting account")
	}

	return a, nil
}

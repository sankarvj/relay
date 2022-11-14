package account

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"go.opencensus.io/trace"
)

var (
	// ErrNotFound is used when a specific User is requested but does not exist.
	ErrNotFound = errors.New("Account not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")
)

var ExistingSubDomains = []string{"www", "app", "csp", "crp", "emp", "client", "clients", "customer", "customers", "employee", "employees"}

// List retrieves a list of existing accounts from the database.
func List(ctx context.Context, accountIDs []string, db *sqlx.DB) ([]Account, error) {
	ctx, span := trace.StartSpan(ctx, "internal.account.List")
	defer span.End()

	accounts := []Account{}
	const q = `SELECT * FROM accounts where account_id = any($1)`

	if err := db.SelectContext(ctx, &accounts, q, pq.Array(accountIDs)); err != nil {
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

func RetrieveByStripeID(ctx context.Context, stripeCusID string, db *sqlx.DB) (*Account, error) {
	ctx, span := trace.StartSpan(ctx, "internal.account.RetrieveByStripeID")
	defer span.End()

	var a Account
	const q = `SELECT * FROM accounts WHERE cus_id = $1`
	if err := db.GetContext(ctx, &a, q, stripeCusID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, errors.Wrapf(err, "selecting account using stripe ID %q", stripeCusID)
	}

	return &a, nil
}

func Delete(ctx context.Context, db *sqlx.DB, accountID string) error {
	ctx, span := trace.StartSpan(ctx, "internal.account.Delete")
	defer span.End()

	const q = `DELETE FROM accounts WHERE account_id = $1`

	if _, err := db.ExecContext(ctx, q, accountID); err != nil {
		return errors.Wrapf(err, "deleting account %s", accountID)
	}

	return nil
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

func UpdateStripeCustomer(ctx context.Context, accountID, cusID, cusEmail string, trailEnds int64, plan int, db *sqlx.DB) error {
	ctx, span := trace.StartSpan(ctx, "internal.account.UpdateStripeCustomer")
	defer span.End()

	expireAt := util.ConvertMilliToTime(trailEnds)

	const q = `UPDATE accounts SET cus_id = $2, cus_mail = $3, expiry = $4, plan = $5
		WHERE account_id = $1`
	_, err := db.ExecContext(ctx, q, accountID,
		cusID, cusEmail, expireAt, plan,
	)
	if err != nil {
		return errors.Wrap(err, "updating account payment plan and stripe customer id")
	}
	return nil
}

func UpdateStripePlan(ctx context.Context, accountID string, plan, seat int, db *sqlx.DB) error {
	ctx, span := trace.StartSpan(ctx, "internal.account.UpdateStripeCustomer")
	defer span.End()

	const q = `UPDATE accounts SET plan = $2, cus_seat = $3
		WHERE account_id = $1`
	_, err := db.ExecContext(ctx, q, accountID,
		plan, seat,
	)
	if err != nil {
		return errors.Wrap(err, "updating account payment plan and seat")
	}
	return nil
}

func UpdateStripeTrailEndDate(ctx context.Context, accountID string, expiry int64, db *sqlx.DB) error {
	ctx, span := trace.StartSpan(ctx, "internal.account.UpdateStripeCustomer")
	defer span.End()

	expireAt := util.ConvertMilliToTime(expiry)

	const q = `UPDATE accounts SET expiry = $2
		WHERE account_id = $1`
	_, err := db.ExecContext(ctx, q, accountID,
		expireAt,
	)
	if err != nil {
		return errors.Wrap(err, "updating account payment plan")
	}
	return nil
}

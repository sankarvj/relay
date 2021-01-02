package account

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
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

// Create inserts a new user into the database. Call AccountBootstrap instead
func Create(ctx context.Context, db *sqlx.DB, n NewAccount, now time.Time) (Account, error) {
	ctx, span := trace.StartSpan(ctx, "internal.account.Create")
	defer span.End()

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

func AccountBootstrap(ctx context.Context, db *sqlx.DB, cu *user.User, accountID, teamID string, n NewAccount, now time.Time) error {
	n.ID = accountID
	a, err := Create(ctx, db, n, now)
	if err != nil {
		return err
	}
	err = user.AddAccounts(ctx, db, cu, a.ID, time.Now())
	if err != nil {
		return errors.Wrap(err, "account inserted but user update failed")
	}

	err = bootstrap.BootstrapTeam(ctx, db, a.ID, teamID, "CRM")
	if err != nil {
		return errors.Wrap(err, "account inserted but team bootstrap failed")
	}

	err = bootstrap.BootstrapUserEntity(ctx, db, cu, a.ID, teamID)
	if err != nil {
		return errors.Wrap(err, "account inserted but users bootstrap failed")
	}
	return nil
}

package handlers

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/account"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
)

var (
	// ErrNotFound is used when a specific User is requested but does not exist.
	ErrNotFound = errors.New("Account not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")
)

// Account represents the Account API method handler set.
type Account struct {
	db            *sqlx.DB
	authenticator *auth.Authenticator
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// List returns all the existing accounts associated with logged-in user
func (a *Account) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Account.List")
	defer span.End()

	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return web.NewShutdownError("claims missing from context")
	}

	accounts, err := account.List(ctx, currentUserID, a.db)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, accounts, http.StatusOK)
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

package handlers

import (
	"context"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/account"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
)

// Account represents the Account API method handler set.
type Account struct {
	db            *sqlx.DB
	authenticator *auth.Authenticator
	rPool         *redis.Pool
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// List returns all the existing accounts associated with logged-in user
func (a *Account) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Account.List")
	defer span.End()

	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		err := errors.New("auth_cliams_missing_from_context") // value used in the UI dont change the string message.
		return web.NewRequestError(err, http.StatusForbidden)
	}

	accounts, err := account.List(ctx, currentUserID, a.db)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, accounts, http.StatusOK)
}

func (a *Account) Availability(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	accName := r.URL.Query().Get("name")
	if accName == "" {
		return web.Respond(ctx, w, false, http.StatusOK)
	}
	_, err := account.CheckAvailability(ctx, accName, a.db)
	if err == account.ErrNotFound {
		return web.Respond(ctx, w, true, http.StatusOK)
	}
	return web.Respond(ctx, w, false, http.StatusOK)
}

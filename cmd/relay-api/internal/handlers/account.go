package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/account"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap"
	"gitlab.com/vjsideprojects/relay/internal/draft"
	"gitlab.com/vjsideprojects/relay/internal/job"
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
		return web.NewShutdownError("auth claims missing from context")
	}

	accounts, err := account.List(ctx, currentUserID, a.db)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, accounts, http.StatusOK)
}

func (a *Account) Availability(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	accName := r.URL.Query().Get("name")
	_, err := account.CheckAvailability(ctx, accName, a.db)
	if err == account.ErrNotFound {
		return web.Respond(ctx, w, true, http.StatusOK)
	}
	return web.Respond(ctx, w, false, http.StatusOK)
}

func (a *Account) Draft(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Account.Draft")
	defer span.End()

	var nd draft.NewDraft
	if err := web.Decode(r, &nd); err != nil {
		return errors.Wrap(err, "")
	}

	draft, err := draft.Create(ctx, nd, time.Now(), a.db)
	if err != nil {
		return errors.Wrapf(err, "Draft: %+v", &draft)
	}

	(&job.Job{}).EventUserSignedUp(draft.AccountName, draft.BusinessEmail, draft.ID, a.db, a.rPool)
	return web.Respond(ctx, w, true, http.StatusCreated)
}

func (a *Account) RetriveDraft(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Account.RetriveDraft")
	defer span.End()

	draftID := params["draft_id"]

	draft, err := draft.Retrieve(ctx, draftID, a.db)
	if err != nil {
		return errors.Wrapf(err, "Draft: %+v", &draft)
	}

	return web.Respond(ctx, w, draft, http.StatusOK)
}

func (a *Account) Launch(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Account.Launch")
	defer span.End()

	draftID := params["draft_id"]
	token := r.URL.Query().Get("token")

	tokenUID, tokenEmail, err := verifyToken(ctx, a.authenticator.FireBaseAdminSDK, token)
	if err != nil {
		return errors.Wrap(err, "verifying token with firebase")
	}

	usr, err := user.RetrieveUserByUniqIdentifier(ctx, a.db, tokenEmail, "")
	if err != nil {
		return errors.Wrapf(err, "retrival of user failed")
	}

	draft, err := draft.Retrieve(ctx, draftID, a.db)
	if err != nil {
		return errors.Wrapf(err, "retrival of draft failed: %+v", &draft)
	}

	tkn, err := authenticate(ctx, tokenEmail, time.Now(), tokenUID, a.authenticator, a.db)
	if err != nil {
		return errors.Wrap(err, "generating token")
	}

	accountID := uuid.New().String()
	nc := account.NewAccount{
		ID:      accountID,
		Name:    draft.AccountName,
		DraftID: draft.ID,
	}

	err = account.Bootstrap(ctx, a.db, a.rPool, &usr, nc, time.Now())
	if err != nil {
		return errors.Wrap(err, "Bootstrap failed")
	}
	//this will take the user to the specific account even multiple accounts exists
	tkn.Accounts = []string{nc.ID}

	//update user with the account
	user.UpdateAccounts(ctx, a.db, &usr, nc.ID, time.Now())

	//TODO: boot crm for now. In future boot based on the launch account selection
	err = bootstrap.BootCRM(accountID, a.db, a.rPool)
	if err != nil {
		return errors.Wrap(err, "Bootstrap CRM failed")
	}

	err = bootstrap.BootCSM(accountID, a.db, a.rPool)
	if err != nil {
		return errors.Wrap(err, "Bootstrap CSM failed")
	}

	return web.Respond(ctx, w, tkn, http.StatusCreated)
}

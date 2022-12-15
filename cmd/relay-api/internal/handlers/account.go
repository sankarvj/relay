package handlers

import (
	"context"
	"net/http"
	"net/mail"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/account"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/token"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
)

// Account represents the Account API method handler set.
type Account struct {
	db            *sqlx.DB
	sdb           *database.SecDB
	authenticator *auth.Authenticator
}

// List returns all the existing accounts associated with logged-in user
func (a *Account) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Account.List")
	defer span.End()

	//for v1/accounts claims subject should have email instead of ID. Check generateUserJWTForAnyAccount
	emailOrUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return err
	}

	users, err := accUsers(ctx, emailOrUserID, a.db)
	if err != nil {
		return err
	}

	accountIDs := make([]string, 0)
	for _, dbUser := range users {
		accountIDs = append(accountIDs, dbUser.AccountID)
	}

	accounts, err := account.List(ctx, accountIDs, a.db)
	if err != nil {
		return err
	}

	vmAccPages := make([]ViewModelAccountPage, 0)
	for _, acc := range accounts {
		vmAccPages = append(vmAccPages, createViewModelAccountPage(acc))
	}

	return web.Respond(ctx, w, vmAccPages, http.StatusOK)
}

func (a *Account) Retrieve(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Account.Retrieve")
	defer span.End()

	account, err := account.Retrieve(ctx, a.db, params["account_id"])
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, createViewModelAccount(account), http.StatusOK)
}

func (a *Account) GenerateToken(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Account.GenerateToken")
	defer span.End()

	account, err := account.Retrieve(ctx, a.db, params["account_id"])
	if err != nil {
		return err
	}

	emailOrUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return err
	}

	users, err := accUsers(ctx, emailOrUserID, a.db)
	if err != nil {
		return err
	}

	var verified bool
	var userEmail string
	for _, dbUser := range users {
		if dbUser.AccountID == account.ID {
			userEmail = dbUser.Email
			verified = true
		}
	}

	if !verified {
		err := errors.New("account_not_associated_with_this_user") // value used in the UI dont change the string message.
		return web.NewRequestError(err, http.StatusForbidden)
	}

	tkn, err := generateUserJWT(ctx, account.ID, userEmail, time.Now(), a.authenticator, a.db)
	if err != nil {
		return web.NewRequestError(err, http.StatusUnauthorized)
	}

	return web.Respond(ctx, w, tkn, http.StatusOK)
}

func (a *Account) APIToken(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Account.APIToken")
	defer span.End()

	account, err := account.Retrieve(ctx, a.db, params["account_id"])
	if err != nil {
		return err
	}

	tkn, err := token.Retrieve(ctx, a.db, params["account_id"])
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, createVMToken(account.Name, *tkn), http.StatusOK)
}

func (a *Account) Availability(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	accName := r.URL.Query().Get("name")
	for _, v := range account.ExistingSubDomains {
		if accName == v {
			return web.Respond(ctx, w, false, http.StatusOK)
		}
	}
	if accName == "" {
		return web.Respond(ctx, w, false, http.StatusOK)
	}
	_, err := account.CheckAvailability(ctx, accName, a.db)
	if err == account.ErrNotFound {
		return web.Respond(ctx, w, true, http.StatusOK)
	}
	return web.Respond(ctx, w, false, http.StatusOK)
}

func (a *Account) RemoveDummyData(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Account.Retrieve")
	defer span.End()

	account, err := account.Retrieve(ctx, a.db, params["account_id"])
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, createViewModelAccount(account), http.StatusOK)
}

func accUsers(ctx context.Context, currentUserEmail string, db *sqlx.DB) ([]user.User, error) {
	var users []user.User
	_, err := mail.ParseAddress(currentUserEmail)
	if err == nil {
		users, err = user.List(ctx, currentUserEmail, "", db)
		if err != nil {
			return users, err
		}
	} else {
		currentUserID := currentUserEmail
		user, err := user.RetrieveUserWOAcc(ctx, db, currentUserID)
		if err != nil {
			return users, err
		}
		users = append(users, *user)
	}

	return users, nil
}

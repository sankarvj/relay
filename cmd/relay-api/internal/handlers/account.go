package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/account"
	"gitlab.com/vjsideprojects/relay/internal/draft"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
	"golang.org/x/crypto/bcrypt"
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

	bmHash, err := bcrypt.GenerateFromPassword([]byte(nd.BusinessEmail), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, "generating business email hash")
	}

	draft, err := draft.Create(ctx, nd, time.Now(), a.db)
	if err != nil {
		return errors.Wrapf(err, "Draft: %+v", &draft)
	}

	encodedHash := base64.StdEncoding.EncodeToString(bmHash)
	magicLink := fmt.Sprintf("home/%s/drafts/%s/identifier/%s", nd.AccountName, draft.ID, encodedHash)
	log.Println("Magic Link --> ", magicLink)

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

	var la account.LaunchAccount
	if err := web.Decode(r, &la); err != nil {
		return errors.Wrap(err, "")
	}

	draft, err := draft.Retrieve(ctx, la.DraftID, a.db)
	if err != nil {
		return errors.Wrapf(err, "retrival of draft failed: %+v", &draft)
	}

	decodedHash, _ := base64.StdEncoding.DecodeString(la.BusinessEmailHash)

	err = bcrypt.CompareHashAndPassword(decodedHash, []byte(draft.BusinessEmail))
	if err != nil { //not equal
		log.Println("err is password comparison ", err)
		return web.Respond(ctx, w, nil, http.StatusForbidden)
	}

	usr, err := user.RetrieveUserByUniqIdentifier(ctx, a.db, draft.BusinessEmail, "")
	if err != nil && err == user.ErrNotFound {
		nu := user.NewUser{
			Name:            la.FirstName,
			Avatar:          util.String(""),
			Email:           draft.BusinessEmail,
			Phone:           util.String(""),
			Provider:        util.String("default"),
			Password:        la.Password,
			PasswordConfirm: la.PasswordConfirm,
			Roles:           []string{auth.RoleAdmin, auth.RoleUser},
		}
		usr, err = user.Create(ctx, a.db, nu, time.Now())
	}
	if err != nil {
		return errors.Wrap(err, "User creation failed")
	}

	log.Println("Authentication----", string(usr.PasswordHash))

	tkn, err := authenticate(ctx, usr.Email, time.Now(), string(usr.PasswordHash), a.authenticator, a.db)
	if err != nil {
		return errors.Wrap(err, "generating token")
	}

	nc := account.NewAccount{
		ID:      uuid.New().String(),
		Name:    la.AccountName,
		Domain:  la.Domain,
		DraftID: draft.ID,
	}

	err = account.Bootstrap(ctx, a.db, a.rPool, &usr, nc, time.Now())
	if err != nil {
		return errors.Wrap(err, "Bootstrap failed")
	}
	//this will take the user to the specific account even multiple accounts exists
	tkn.Accounts = []string{nc.ID}

	return web.Respond(ctx, w, tkn, http.StatusCreated)
}

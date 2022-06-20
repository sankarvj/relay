package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/account"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap"
	"gitlab.com/vjsideprojects/relay/internal/draft"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
)

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
	fbToken := r.URL.Query().Get("token")    // firebase token
	mlToken := r.URL.Query().Get("ml_token") // magiclink token

	_, tokenEmail, err := verifyToken(ctx, a.authenticator.FireBaseAdminSDK, fbToken)
	if err != nil {
		return errors.Wrap(err, "verifying fbToken")
	}

	userInfo, err := auth.AuthenticateToken(mlToken, a.rPool)
	if err != nil {
		return errors.Wrap(err, "verifying mlToken")
	}

	if userInfo.Email != tokenEmail {
		return errors.Wrap(err, "token mismatch detected")
	}

	// all authentication completed. Proceed with the next steps
	usr, err := user.RetrieveUserByUniqIdentifier(ctx, a.db, tokenEmail, "")
	if err == user.ErrNotFound {
		usr, err = createNewVerifiedUser(ctx, util.NameInEmail(userInfo.Email), userInfo.Email, []string{auth.RoleAdmin}, a.db)
		if err != nil {
			return errors.Wrapf(err, "creating new user failed. please contact support@workbaseone.com")
		}
	} else if err != nil {
		return errors.Wrapf(err, "retrival of user failed")
	}

	dft, err := draft.Retrieve(ctx, draftID, a.db)
	if err != nil {
		return errors.Wrapf(err, "retrival of draft failed")
	}

	accountID := uuid.New().String()
	nc := account.NewAccount{
		ID:      accountID,
		Name:    dft.AccountName,
		DraftID: dft.ID,
	}

	acc, err := account.Create(ctx, a.db, nc, time.Now())
	if err != nil {
		return err
	}

	tkn, err := generateJWT(ctx, tokenEmail, time.Now(), a.authenticator, a.db)
	if err != nil {
		return errors.Wrap(err, "generating token")
	}
	//this will take the user in the frontend to the specific account even multiple accounts exists
	tkn.Accounts = []string{nc.ID}

	err = bootstrap.Bootstrap(ctx, a.db, a.rPool, a.authenticator.FireBaseAdminSDK, acc.ID, &usr)
	if err != nil {
		return errors.Wrap(err, "core bootstrap failed")
	}

	if util.Contains(dft.Teams, "crm") {
		err = bootstrap.BootCRM(accountID, usr.ID, a.db, a.rPool, a.authenticator.FireBaseAdminSDK)
		if err != nil {
			return errors.Wrap(err, "CRM bootstrap  failed")
		}
	}

	if util.Contains(dft.Teams, "csm") {
		err = bootstrap.BootCSM(accountID, usr.ID, a.db, a.rPool, a.authenticator.FireBaseAdminSDK)
		if err != nil {
			return errors.Wrap(err, "CSM bootstrap failed")
		}
	}

	draft.Delete(ctx, draftID, a.db)

	return web.Respond(ctx, w, tkn, http.StatusCreated)
}

func createNewVerifiedUser(ctx context.Context, name, email string, roles []string, db *sqlx.DB) (user.User, error) {
	nu := user.NewUser{
		Name:            name,
		Avatar:          util.String(""),
		Email:           email,
		Phone:           util.String(""),
		Provider:        util.String("default"),
		Password:        "",
		PasswordConfirm: "",
		Verified:        true,
		Roles:           roles,
	}
	u, err := user.Create(ctx, db, nu, time.Now())
	return u, err
}

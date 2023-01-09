package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/account"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap"
	"gitlab.com/vjsideprojects/relay/internal/draft"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/notification"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/database/dbservice"
	"gitlab.com/vjsideprojects/relay/internal/platform/stream"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/team"
	"gitlab.com/vjsideprojects/relay/internal/token"
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

	(&job.Job{}).EventUserSignedUp(draft.AccountName, draft.BusinessEmail, draft.ID, a.db, a.sdb)
	return web.Respond(ctx, w, true, http.StatusCreated)
}

func (a *Account) Launch(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Account.Launch")
	defer span.End()

	draftID := params["draft_id"]
	fbToken := r.URL.Query().Get("token")    // firebase token
	mlToken := r.URL.Query().Get("ml_token") // magiclink token

	_, tokenEmail, err := verifyToken(ctx, a.authenticator.FireBaseAdminSDK, fbToken)
	if err != nil {
		return web.NewRequestError(errors.New("User firebase token mismatch"), http.StatusUnauthorized)
	}

	userInfo, err := auth.AuthenticateToken(mlToken, a.sdb)
	if err != nil {
		return web.NewRequestError(errors.New("User magic link token mismatch"), http.StatusUnauthorized)
	}

	if userInfo.Email != tokenEmail {
		//TODO: it seems the token is compromised. Remove the token
		return web.NewRequestError(errors.New("User magiclink token mismatch. Token invalidated"), http.StatusUnauthorized)
	}

	dft, err := draft.Retrieve(ctx, draftID, a.db)
	if err != nil {
		return web.NewRequestError(errors.Wrap(err, "Draft not found"), http.StatusInternalServerError)
	}

	accountID := uuid.New().String()
	nc := account.NewAccount{
		ID:             accountID,
		Name:           dft.AccountName,
		Domain:         notification.Hostname(dft.AccountName, dft.Host),
		DraftID:        dft.ID,
		CustomerStatus: account.StatusTrial,
		CustomerPlan:   account.PlanPro,
		TrailStart:     util.GetMilliSecondsFloatReduced(time.Now()), //webhooks from stripe will update the values anyways
		TrailEnd:       util.AddMilliSecondsFloat(time.Now(), 13),    //webhooks from stripe will update the values anyways
		UseDB:          string(dbservice.Bee),
	}

	acc, err := account.Create(ctx, a.db, nc, time.Now())
	if err != nil {
		return web.NewRequestError(errors.Wrap(err, "Account creation failed"), http.StatusInternalServerError)
	}

	// all authentication completed. Proceed with the next steps
	usr, err := user.RetrieveUserByUniqIdentifier(ctx, accountID, tokenEmail, "", a.db)
	if err == user.ErrNotFound {
		usr, err = createNewVerifiedUser(ctx, accountID, util.NameInEmail(userInfo.Email), userInfo.Email, []string{auth.RoleAdmin}, a.db)
		if err != nil {
			return web.NewRequestError(errors.Wrap(err, "creating new user failed. please contact support@workbaseone.com"), http.StatusUnauthorized)
		}
	} else if err != nil {
		return web.NewRequestError(errors.Wrap(err, "retrival of user failed. please contact support@workbaseone.com"), http.StatusUnauthorized)
	}

	systemToken, err := generateSystemUserJWT(ctx, acc.ID, []string{}, time.Now(), a.authenticator, a.db)
	if err != nil {
		account.Delete(ctx, a.db, acc.ID)
		return web.NewRequestError(errors.Wrap(err, "System JWT creation failed"), http.StatusInternalServerError)
	}
	err = token.Create(ctx, a.db, systemToken, acc.ID, time.Now())
	if err != nil {
		account.Delete(ctx, a.db, acc.ID)
		return web.NewRequestError(errors.Wrap(err, "System JWT token save failed"), http.StatusInternalServerError)
	}

	userToken, err := generateUserJWT(ctx, acc.ID, tokenEmail, time.Now(), a.authenticator, a.db)
	if err != nil {
		account.Delete(ctx, a.db, acc.ID)
		return web.NewRequestError(errors.Wrap(err, "User JWT creation failed"), http.StatusInternalServerError)
	}
	//this will take the user in the frontend to the specific account even multiple accounts exists
	userToken.Accounts = []string{acc.ID}

	err = bootstrap.Bootstrap(ctx, a.db, a.sdb, a.authenticator.FireBaseAdminSDK, acc.ID, acc.Name, &usr)
	if err != nil {
		account.Delete(ctx, a.db, acc.ID)
		return web.NewRequestError(errors.Wrap(err, "Cannot bootstrap your account. Please contact support"), http.StatusInternalServerError)
	}
	//this will show onboard page in the UI
	userToken.JustLaunched = true

	go a.bootApp(ctx, acc.ID, usr.ID, dft)
	//FIX THIS
	go job.NewJob(a.db, a.sdb, a.authenticator.FireBaseAdminSDK).Stream(stream.NewAccountLaunchMessage(ctx, a.db, accountID, usr.ID))

	return web.Respond(ctx, w, userToken, http.StatusCreated)
}

func (a *Account) bootApp(ctx context.Context, accountID, userID string, dft *draft.Draft) error {
	if util.Contains(dft.Teams, team.PredefinedTeamCRP) {
		err := bootstrap.BootCRM(accountID, userID, a.db, a.sdb, a.authenticator.FireBaseAdminSDK)
		if err != nil {
			account.Delete(ctx, a.db, accountID)
			return web.NewRequestError(errors.Wrap(err, "Cannot bootstrap your account. Please contact support"), http.StatusInternalServerError)
		}
	}

	if util.Contains(dft.Teams, team.PredefinedTeamCSP) {
		err := bootstrap.BootCSM(accountID, userID, a.db, a.sdb, a.authenticator.FireBaseAdminSDK)
		if err != nil {
			account.Delete(ctx, a.db, accountID)
			return web.NewRequestError(errors.Wrap(err, "Cannot bootstrap your account. Please contact support"), http.StatusInternalServerError)
		}
	}

	if util.Contains(dft.Teams, team.PredefinedTeamEMP) {
		err := bootstrap.BootEM(accountID, userID, a.db, a.sdb, a.authenticator.FireBaseAdminSDK)
		if err != nil {
			account.Delete(ctx, a.db, accountID)
			return web.NewRequestError(errors.Wrap(err, "Cannot bootstrap your account. Please contact support"), http.StatusInternalServerError)
		}
	}
	draft.Delete(ctx, dft.ID, a.db)
	return nil
}

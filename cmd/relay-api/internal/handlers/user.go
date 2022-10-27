package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	firebase "firebase.google.com/go"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/stream"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"gitlab.com/vjsideprojects/relay/internal/visitor"
	"go.opencensus.io/trace"
	"google.golang.org/api/option"
)

const (
	NoEntityID = "00000000-0000-0000-0000-000000000000"
)

var ErrForbiddenMLToken = web.NewRequestError(
	errors.New("Magic link token expired or not found. Please ask your admin to reinvite"),
	http.StatusForbidden,
)

// User represents the User API method handler set.
type User struct {
	db            *sqlx.DB
	sdb           *database.SecDB
	authenticator *auth.Authenticator
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// List returns all the existing users in the system.
func (u *User) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.List")
	defer span.End()

	users, err := user.List(ctx, u.db)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, users, http.StatusOK)
}

// Retrieve returns the specified user from the system.
func (u *User) Retrieve(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.Retrieve")
	defer span.End()

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	usr, err := user.Retrieve(ctx, claims, u.db)
	if err != nil {
		switch err {
		case user.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case user.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case user.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "Id: %s", params["id"])
		}
	}

	return web.Respond(ctx, w, createViewModelUser(*usr, ""), http.StatusOK)
}

// Create inserts a new user into the system.
func (u *User) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.Create")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	var nu user.NewUser
	if err := web.Decode(r, &nu); err != nil {
		return errors.Wrap(err, "")
	}

	usr, err := user.RetrieveUserByUniqIdentifier(ctx, u.db, nu.Email, *nu.Phone)

	if err != nil && err == user.ErrNotFound {
		usr, err = user.Create(ctx, u.db, nu, v.Now)
		if err != nil {
			return errors.Wrapf(err, "User: %+v created", &usr)
		}

	} else if err == nil {
		err := user.UpdatePassword(ctx, u.db, usr.ID, nu.Password, v.Now)
		if err != nil {
			return errors.Wrapf(err, "User: %+v password updated", &usr)
		}
	}

	return web.Respond(ctx, w, usr, http.StatusCreated)
}

// Update updates the specified user in the system.
func (u *User) Update(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.Update")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return err
	}

	var upd user.UpdateUser
	if err := web.Decode(r, &upd); err != nil {
		return errors.Wrap(err, "")
	}

	err = user.Update(ctx, claims, u.db, currentUserID, upd, v.Now)
	if err != nil {
		switch err {
		case user.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case user.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case user.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "ID: %s  User: %+v", params["id"], &upd)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// Delete removes the specified user from the system.
func (u *User) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.Delete")
	defer span.End()

	err := user.Delete(ctx, u.db, params["id"])
	if err != nil {
		switch err {
		case user.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case user.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case user.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "Id: %s", params["id"])
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// Token handles a request to authenticate a user. It expects a request using
// Code and Provider
func (u *User) Verfiy(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.Token")
	defer span.End()

	token := r.URL.Query().Get("token")
	mltoken := r.URL.Query().Get("ml_token")
	if mltoken != "" {
		return u.MLVerify(ctx, w, r, params)
	}

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	_, tokenEmail, err := verifyToken(ctx, u.authenticator.FireBaseAdminSDK, token)
	if err != nil {
		return errors.Wrap(err, "verifying token with firebase")
	}

	tkn, err := generateUserJWT(ctx, tokenEmail, v.Now, u.authenticator, u.db)
	if err != nil {
		return web.NewRequestError(err, http.StatusUnauthorized)
	}

	return web.Respond(ctx, w, tkn, http.StatusOK)
}

func (u *User) MLVerify(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.MLTokenVerification")
	defer span.End()

	token := r.URL.Query().Get("ml_token")

	userInfo, err := auth.AuthenticateToken(token, u.sdb)
	if err != nil {
		log.Println("***> unexpected error occurred in MLVerify. error: ", err)
		return ErrForbiddenMLToken
	}
	_, err = user.RetrieveUserByUniqIdentifier(ctx, u.db, userInfo.Email, "")
	if err != nil {
		userInfo.Verified = false
	}
	return web.Respond(ctx, w, userInfo, http.StatusOK)
}

func (u *User) Join(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.TokenLink")
	defer span.End()

	fbToken := r.URL.Query().Get("token")    // firebase token
	mlToken := r.URL.Query().Get("ml_token") // magiclink token

	_, tokenEmail, err := verifyToken(ctx, u.authenticator.FireBaseAdminSDK, fbToken)
	if err != nil {
		return errors.Wrap(err, "verifying fbToken")
	}

	userInfo, err := auth.AuthenticateToken(mlToken, u.sdb)
	if err != nil {
		return web.NewRequestError(errors.New("User doest not exist. Please create a account first"), http.StatusUnauthorized)
	}

	if userInfo.Email != tokenEmail {
		return web.NewRequestError(errors.Wrap(err, "token mismatch detected"), http.StatusUnauthorized)
	}

	// all authentication completed. Proceed with the next steps
	usr, err := user.RetrieveUserByUniqIdentifier(ctx, u.db, tokenEmail, "")
	if err == user.ErrNotFound {
		usr, err = createNewVerifiedUser(ctx, util.NameInEmail(userInfo.Email), userInfo.Email, []string{auth.RoleAdmin}, u.db)
		if err != nil {
			return errors.Wrapf(err, "creating new user failed. please contact support@workbaseone.com")
		}
	} else if err != nil {
		return errors.Wrapf(err, "retrival of user failed for reason other than not found")
	}

	err = usr.UpdateAccounts(ctx, u.db, usr.AddAccount(userInfo.AccountID, userInfo.MemberID))
	if err != nil {
		return errors.Wrap(err, "adding accounts to user failed")
	}

	err = user.SetAsVerified(ctx, u.db, &usr, time.Now())
	if err != nil {
		return errors.Wrap(err, "setting user verified failed")
	}

	//add newly created userID to the members
	err = u.updateMemberUserID(ctx, userInfo.AccountID, userInfo.MemberID, usr.ID)
	if err != nil {
		return errors.Wrap(err, "adding member record for this user failed")
	}

	tkn, err := generateUserJWT(ctx, tokenEmail, time.Now(), u.authenticator, u.db)
	if err != nil {
		return errors.Wrap(err, "generating token")
	}
	//this will take the user in the frontend to the specific account even multiple accounts exists
	tkn.Accounts = []string{userInfo.AccountID}

	return web.Respond(ctx, w, tkn, http.StatusCreated)
}

func (u *User) Visit(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.Visit")
	defer span.End()

	mlToken := r.URL.Query().Get("ml_token") // magiclink token

	userInfo, err := auth.AuthenticateToken(mlToken, u.sdb)
	if err != nil {
		return errors.Wrap(err, "verifying mlToken")
	}

	vis, err := visitor.Retrieve(ctx, userInfo.AccountID, userInfo.MemberID, u.db)
	if err != nil {
		return err
	}

	if vis.Token != mlToken {
		return errors.Wrap(err, "token mismatch detected")
	}

	_, err = user.RetrieveUserByUniqIdentifier(ctx, u.db, vis.Email, "")
	if err == user.ErrNotFound {
		_, err = createNewVerifiedUser(ctx, util.NameInEmail(userInfo.Email), userInfo.Email, []string{auth.RoleVisitor}, u.db)
		if err != nil {
			return errors.Wrapf(err, "creating new user failed. please contact support@workbaseone.com")
		}
	} else if err != nil {
		return errors.Wrapf(err, "retrival of user failed for reason other than not found")
	}

	tkn, err := generateUserJWT(ctx, vis.Email, time.Now(), u.authenticator, u.db)
	if err != nil {
		return errors.Wrap(err, "generating token")
	}
	//this will take the user in the frontend to the specific account even multiple accounts exists
	tkn.Accounts = []string{vis.AccountID}
	tkn.Team = vis.TeamID
	tkn.Entity = vis.EntityID
	tkn.Item = vis.ItemID

	return web.Respond(ctx, w, tkn, http.StatusCreated)
}

func (u *User) updateMemberUserID(ctx context.Context, accountID, memberID, userID string) error {
	ownerEntity, err := entity.RetrieveFixedEntity(ctx, u.db, accountID, "", entity.FixedEntityOwner)
	if err != nil {
		return err
	}
	existingItem, err := item.Retrieve(ctx, ownerEntity.ID, memberID, u.db)
	if err != nil {
		return err
	}

	existingFields := existingItem.Fields()
	updatedFields := make(map[string]interface{}, 0)
	namedKeys := entity.NamedKeysMap(ownerEntity.FieldsIgnoreError())
	for name, k := range namedKeys {
		updatedFields[k] = existingFields[k]
		if name == "user_id" {
			updatedFields[k] = userID
		}
	}

	it, err := item.UpdateFields(ctx, u.db, ownerEntity.ID, existingItem.ID, updatedFields)
	if err != nil {
		return err
	}

	//stream
	go job.NewJob(u.db, u.sdb, u.authenticator.FireBaseAdminSDK).Stream(stream.NewUpdateItemMessage(ctx, u.db, accountID, userID, ownerEntity.ID, existingItem.ID, it.Fields(), existingItem.Fields()))

	//adding in the members items for reverse lookup of userID from memberID.
	return nil
}

func verifyToken(ctx context.Context, adminSDK string, idToken string) (string, string, error) {
	opt := option.WithCredentialsFile(adminSDK)
	// Initialize default app
	// config := &firebase.Config{ProjectID: "relay-94b69"}
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return "", "", errors.Wrap(err, "")
	}

	// Access auth service from the default app
	client, err := app.Auth(ctx)
	if err != nil {
		return "", "", errors.Wrap(err, "")
	}

	token, err := client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return "", "", errors.Wrap(err, "verifying token with firebase")
	}
	return token.UID, token.Claims["email"].(string), nil
}

func generateUserJWT(ctx context.Context, email string, now time.Time, a *auth.Authenticator, db *sqlx.DB) (*UserToken, error) {
	dbUser, err := user.RetrieveUserByUniqIdentifier(ctx, db, email, "")
	if err != nil {
		errors.Wrap(err, "user does not exist in the DB. Cannot generate JWT token")
		return nil, web.NewRequestError(err, http.StatusUnauthorized)
	}
	claims := auth.NewClaims(dbUser.ID, dbUser.Roles, now, 96*time.Hour)
	if err != nil {
		switch err {
		case user.ErrAuthenticationFailure:
			return nil, web.NewRequestError(err, http.StatusUnauthorized)
		default:
			return nil, web.NewRequestError(err, http.StatusUnauthorized)
		}
	}

	var tkn UserToken
	tkn.Token, err = a.GenerateToken(claims)
	if err != nil {
		return nil, err
	}
	tkn.Accounts = dbUser.AccountIDs()
	return &tkn, nil
}

func generateSystemUserJWT(ctx context.Context, accountID string, scope []string, now time.Time, a *auth.Authenticator, db *sqlx.DB) (string, error) {
	claims := auth.NewClaims(accountID, scope, now, 24*7*1000*time.Hour)

	systemToken, err := a.GenerateToken(claims)
	if err != nil {
		return "", err
	}
	return systemToken, nil
}

func userMap(ctx context.Context, userIDs map[string]bool, db *sqlx.DB) (map[string]*user.User, error) {
	userMap := make(map[string]*user.User, 0)

	users, err := user.BulkRetrieveUsers(ctx, userkeys(userIDs), db)
	if err != nil {
		return userMap, err
	}

	for _, u := range users {
		userMap[u.ID] = &u
	}
	userMap[user.UUID_SYSTEM_USER] = &user.User{
		ID:     user.UUID_SYSTEM_USER,
		Name:   util.String("System"),
		Avatar: util.String("https://avatars.dicebear.com/api/bottts/system.svg"),
	}
	userMap[user.UUID_ENGINE_USER] = &user.User{
		ID:     user.UUID_ENGINE_USER,
		Name:   util.String("Automation"),
		Avatar: util.String("https://avatars.dicebear.com/api/bottts/workflow.svg"),
	}
	userMap[user.UUID_ANONYMOUS_USER] = &user.User{
		ID:     user.UUID_ANONYMOUS_USER,
		Name:   util.String("Anonymous"),
		Avatar: util.String("https://avatars.dicebear.com/api/bottts/anonymous.svg"),
	}
	return userMap, nil
}

func userkeys(oneMap map[string]bool) []string {
	keys := make([]string, 0, len(oneMap))
	for k := range oneMap {
		keys = append(keys, k)
	}
	return keys
}

func userAvatarNameEmail(u *user.User) (string, string, string) {
	if u != nil {
		return *u.Avatar, *u.Name, u.Email
	}
	return "", "", ""
}

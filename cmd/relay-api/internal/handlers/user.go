package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	firebase "firebase.google.com/go"
	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
	"google.golang.org/api/option"
)

const (
	NoEntityID = "00000000-0000-0000-0000-000000000000"
)

// User represents the User API method handler set.
type User struct {
	db            *sqlx.DB
	rPool         *redis.Pool
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

	return web.Respond(ctx, w, createViewModelUser(*usr), http.StatusOK)
}

func (u *User) Invite(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.Invite")
	defer span.End()

	accountID := params["account_id"]

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	// acc, err := account.Retrieve(ctx, u.db, accountID)
	// if err != nil {
	// 	return err
	// }

	// currentUser, err := user.RetrieveCurrentUser(ctx, u.db)
	// if err != nil {
	// 	return err
	// }

	var nusers []user.NewUser
	if err := web.Decode(r, &nusers); err != nil {
		return errors.Wrap(err, "")
	}

	//TODO add validation to restrict inviting users upto 100

	var users []user.User
	for _, nu := range nusers {
		nu.Password = ""        //safty
		nu.PasswordConfirm = "" //safty
		nu.AccountIDs = []string{accountID}
		usr, err := user.RetrieveUserByUniqIdentifier(ctx, u.db, nu.Email, *nu.Phone)
		if err != nil {
			if err == user.ErrNotFound {
				usr, err = user.Create(ctx, u.db, nu, v.Now)
				if err != nil {
					log.Println("***> unexpected error when creating new users to the account. error: ", err)
					usr.ID = "" //symbolically telling the UI that the invitation for the user is failed.
				}
			} else {
				log.Println("***> unexpected error when retriving users when inviting. error: ", err)
			}
		} else { //TODO update account ID

		}
		users = append(users, usr)

		if usr.ID != "" {
			//TODO push this to stream/queue
		}

	}

	return web.Respond(ctx, w, users, http.StatusCreated)
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

	var upd user.UpdateUser
	if err := web.Decode(r, &upd); err != nil {
		return errors.Wrap(err, "")
	}

	err := user.Update(ctx, claims, u.db, params["id"], upd, v.Now)
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

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	tokenUID, tokenEmail, err := verifyToken(ctx, u.authenticator.FireBaseAdminSDK, token)
	if err != nil {
		return errors.Wrap(err, "verifying token with firebase")
	}

	tkn, err := authenticate(ctx, tokenEmail, v.Now, tokenUID, u.authenticator, u.db)
	if err != nil {
		return errors.Wrap(err, "generating token")
	}

	return web.Respond(ctx, w, tkn, http.StatusOK)
}

func (u *User) Join(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.TokenLink")
	defer span.End()

	token := r.URL.Query().Get("token")

	userInfo, err := auth.AuthenticateToken(token, u.rPool)
	if err != nil {
		return errors.Wrap(err, "authrnticating token")
	}

	log.Printf("userInfo--%+v", userInfo)

	usr, err := user.RetrieveUserByUniqIdentifier(ctx, u.db, userInfo.Email, "")
	if err != nil && err == user.ErrNotFound {
		userInfo.NewUser = true
		nu := user.NewUser{
			Name:            util.NameInEmail(userInfo.Email),
			Avatar:          util.String(""),
			Email:           userInfo.Email,
			Phone:           util.String(""),
			Provider:        util.String("default"),
			Password:        "",
			PasswordConfirm: "",
			AccountIDs:      []string{userInfo.AccountID},
			Roles:           []string{auth.RoleUser},
		}
		usr, err = user.Create(ctx, u.db, nu, time.Now())
	} else {
		userInfo.NewUser = false
		err = user.UpdateAccounts(ctx, u.db, &usr, userInfo.AccountID, time.Now())
	}
	if err != nil {
		return errors.Wrap(err, "User creation failed")
	}

	//add newly created userID to the members
	err = updateMemberUserID(ctx, userInfo.AccountID, userInfo.MemberID, usr.ID, u.db)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, userInfo, http.StatusOK)
}

func (u *User) Launch(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.TokenLink")
	defer span.End()

	token := r.URL.Query().Get("token")

	userInfo, err := auth.AuthenticateToken(token, u.rPool)
	if err != nil {
		return errors.Wrap(err, "authrnticating token")
	}

	_, err = user.RetrieveUserByUniqIdentifier(ctx, u.db, userInfo.Email, "")
	if err != nil && err == user.ErrNotFound {
		userInfo.NewUser = true
		nu := user.NewUser{
			Name:            util.NameInEmail(userInfo.Email),
			Avatar:          util.String(""),
			Email:           userInfo.Email,
			Phone:           util.String(""),
			Provider:        util.String("default"),
			Password:        "",
			PasswordConfirm: "",
			Roles:           []string{auth.RoleUser, auth.RoleAdmin},
		}
		_, err = user.Create(ctx, u.db, nu, time.Now())
	} else {
		userInfo.NewUser = false
	}
	if err != nil {
		return errors.Wrap(err, "User creation failed")
	}

	//add newly created userID to the members
	// err = updateMemberUserID(ctx, userInfo.AccountID, userInfo.TeamID, userInfo.MemberID, usr.ID, u.db)
	// if err != nil {
	// 	return err
	// }

	return web.Respond(ctx, w, userInfo, http.StatusOK)
}

func updateMemberUserID(ctx context.Context, accountID, memberID, userID string, db *sqlx.DB) error {
	ownerEntity, err := entity.RetrieveFixedEntity(ctx, db, accountID, "", entity.FixedEntityOwner)
	if err != nil {
		return err
	}
	var userEntityItem entity.UserEntity
	valueAddedFields, updateFunc, err := entity.RetrieveFixedItem(ctx, accountID, ownerEntity.ID, memberID, db)
	if err != nil {
		return err
	}
	err = entity.ParseFixedEntity(valueAddedFields, &userEntityItem)
	if err != nil {
		return err
	}
	userEntityItem.UserID = userID
	//adding in the members items for reverse lookup of userID from memberID.
	return updateFunc(ctx, userEntityItem, db)
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

func authenticate(ctx context.Context, email string, now time.Time, uid string, a *auth.Authenticator, db *sqlx.DB) (*UserToken, error) {
	log.Println("email", email)
	log.Println("uid", uid)

	dbUser, claims, err := user.Authenticate(ctx, db, now, email, uid)
	if err != nil {
		switch err {
		case user.ErrAuthenticationFailure:
			return nil, web.NewRequestError(err, http.StatusUnauthorized)
		default:
			return nil, errors.Wrap(err, "authenticating")
		}
	}

	var tkn UserToken
	tkn.Token, err = a.GenerateToken(claims)
	if err != nil {
		return nil, err
	}
	tkn.Accounts = dbUser.AccountIDs
	return &tkn, nil
}

func createViewModelUser(u user.User) user.ViewModelUser {
	return user.ViewModelUser{
		Name:      *u.Name,
		Avatar:    *u.Avatar,
		Email:     u.Email,
		Phone:     *u.Phone,
		Roles:     u.Roles,
		CreatedAt: u.CreatedAt.String(),
	}
}

type UserToken struct {
	Token    string   `json:"token"`
	Accounts []string `json:"accounts"`
}

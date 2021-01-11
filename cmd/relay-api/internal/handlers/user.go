package handlers

import (
	"context"
	"log"
	"net/http"

	firebase "firebase.google.com/go"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
	"google.golang.org/api/option"
)

// User represents the User API method handler set.
type User struct {
	db            *sqlx.DB
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

	usr, err := user.Create(ctx, u.db, nu, v.Now)
	if err != nil {
		return errors.Wrapf(err, "User: %+v", &usr)
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
func (u *User) Token(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.Token")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	opt := option.WithCredentialsFile(u.authenticator.FireBaseAdminSDK)
	// Initialize default app
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return errors.Wrap(err, "")
	}

	// Access auth service from the default app
	client, err := app.Auth(context.Background())
	if err != nil {
		return errors.Wrap(err, "")
	}

	token, err := client.VerifyIDToken(ctx, params["id"])
	if err != nil {
		return errors.Wrap(err, "verifying token with firebase")
	}

	userRecord, err := client.GetUser(ctx, token.UID)
	if err != nil {
		return errors.Wrap(err, "fetching user from token UID")
	}

	log.Printf("sk/saravana please replace the word sk_replacetokenhere/sarvana_replacetokenhere in seed.go with this token to login %s", token.UID)

	dbUser, claims, err := user.Authenticate(ctx, u.db, v.Now, userRecord.Email, token.UID)
	if err != nil {
		switch err {
		case user.ErrAuthenticationFailure:
			return web.NewRequestError(err, http.StatusUnauthorized)
		default:
			return errors.Wrap(err, "authenticating")
		}
	}

	var tkn struct {
		Token    string   `json:"token"`
		Accounts []string `json:"accounts"`
	}
	tkn.Token, err = u.authenticator.GenerateToken(claims)
	tkn.Accounts = dbUser.AccountIDs
	if err != nil {
		return errors.Wrap(err, "generating token")
	}

	//dbuser, err := model.CreateNewUserIfNotExists(name, email, phone, avatar, provider, uid, token.Expires, token.IssuedAt, emailVerified)
	if err != nil {
		return errors.Wrap(err, "")
	}

	return web.Respond(ctx, w, tkn, http.StatusOK)
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

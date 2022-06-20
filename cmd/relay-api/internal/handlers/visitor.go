package handlers

import (
	"context"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/visitor"
	"go.opencensus.io/trace"
)

// Visitor represents the Visitor API method handler set.
type Visitor struct {
	db            *sqlx.DB
	rPool         *redis.Pool
	authenticator *auth.Authenticator
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// List returns all the existing items
func (v *Visitor) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Visitor.List")
	defer span.End()

	mlToken := r.URL.Query().Get("ml_token")

	userInfo, err := auth.AuthenticateToken(mlToken, v.rPool)
	if err != nil {
		return errors.Wrap(err, "verifying mlToken")
	}

	vis, err := visitor.Retrieve(ctx, userInfo.AccountID, userInfo.MemberID, v.db)
	if err != nil {
		return err
	}

	if vis.Token != mlToken {
		return errors.Wrap(err, "token mismatch detected")
	}

	params["account_id"] = vis.AccountID
	params["team_id"] = vis.TeamID
	params["entity_id"] = vis.EntityID
	params["item_id"] = vis.ItemID

	i := Item{
		db:            v.db,
		rPool:         v.rPool,
		authenticator: v.authenticator,
	}

	return web.Respond(ctx, w, i.List(ctx, w, r, params), http.StatusOK)
}

//Update updates the item
func (v *Visitor) Update(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Visitor.Update")
	defer span.End()
	response := ""
	return web.Respond(ctx, w, response, http.StatusOK)
}

// Create inserts a new team into the system.
func (v *Visitor) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Visitor.Create")
	defer span.End()
	response := ""
	return web.Respond(ctx, w, response, http.StatusCreated)
}

// Retrieve gets the specified item with field meta from the database.
func (v *Visitor) Retrieve(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Visitor.Retrieve")
	defer span.End()
	response := ""
	return web.Respond(ctx, w, response, http.StatusOK)
}

func (v *Visitor) ListRelations(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Visitor.List")
	defer span.End()
	response := ""
	return web.Respond(ctx, w, response, http.StatusOK)
}

func (v *Visitor) ChildItems(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Visitor.ChildItems")
	defer span.End()
	response := ""
	return web.Respond(ctx, w, response, http.StatusOK)
}

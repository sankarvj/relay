package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/team"
	"go.opencensus.io/trace"
)

// Team represents the Team API method handler set.
type Team struct {
	db            *sqlx.DB
	authenticator *auth.Authenticator
}

// List returns all the existing accounts associated with logged-in user
func (t *Team) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Team.List")
	defer span.End()

	teams, err := team.List(ctx, params["account_id"], t.db)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, teams, http.StatusOK)
}

// Create inserts a new team into the system.
func (t *Team) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Team.Create")
	defer span.End()

	var nt team.NewTeam
	if err := web.Decode(r, &nt); err != nil {
		return errors.Wrap(err, "")
	}

	//set account_id from the request path
	nt.AccountID = params["account_id"]
	team, err := team.Create(ctx, t.db, nt, time.Now())
	if err != nil {
		return errors.Wrapf(err, "Team: %+v", &team)
	}

	return web.Respond(ctx, w, team, http.StatusCreated)
}

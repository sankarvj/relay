package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/notification"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"go.opencensus.io/trace"
)

type Notification struct {
	db            *sqlx.DB
	authenticator *auth.Authenticator
	rPool         *redis.Pool
}

func (n *Notification) Register(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Notification.Register")
	defer span.End()

	accountID := params["account_id"]

	var cr notification.ClientRegister
	if err := web.Decode(r, &cr); err != nil {
		return errors.Wrap(err, "")
	}

	currentUserID, err := user.RetrieveCurrentUserID(ctx)
	if err != nil {
		return web.NewShutdownError("auth claims missing from context")
	}
	cr.AccountID = accountID
	cr.UserID = currentUserID
	_, err = notification.CreateClient(ctx, n.db, cr, time.Now())
	if err != nil {
		return errors.Wrapf(err, "failure in saving the client token")
	}

	return web.Respond(ctx, w, true, http.StatusOK)
}

package handlers

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/slack"
	"go.opencensus.io/trace"
)

// Event provides support for slack events handler
type Event struct {
	db            *sqlx.DB
	authenticator *auth.Authenticator
	// ADD OTHER STATE LIKE THE LOGGER IF NEEDED.
}

func (e *Event) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.event.Create")
	defer span.End()
	// body, _ := ioutil.ReadAll(r.Body)
	// fmt.Println("BODY :: ", string(body))

	var sp slack.Payload
	if err := web.DecodeAllowUnknown(r, &sp); err != nil {
		return errors.Wrap(err, "")
	}

	s := slack.Slack{
		BotToken: e.authenticator.SlackBotToken,
		Payload:  sp,
	}

	if err := s.Call(); err != nil {
		return web.NewRequestError(err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, nil, http.StatusOK)
}

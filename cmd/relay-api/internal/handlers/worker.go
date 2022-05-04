package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/stream"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"go.opencensus.io/trace"
)

type Worker struct {
	db            *sqlx.DB
	rPool         *redis.Pool
	authenticator *auth.Authenticator
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// List returns all the existing users in the system.
func (wrk *Worker) receiveSQSPayload(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.worker.receiveSQSPayload")
	defer span.End()

	var message stream.Message
	if err := web.Decode(r, &message); err != nil {
		return errors.Wrap(err, "")
	}

	log.Println("Received Message ---> ", message)
	err := job.NewJob(wrk.db, wrk.rPool).Post(&message)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, message, http.StatusOK)
}

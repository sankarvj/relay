package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"go.opencensus.io/trace"
)

// Flow represents the journey
type Flow struct {
	db            *sqlx.DB
	rPool         *redis.Pool
	authenticator *auth.Authenticator
}

// List returns all the existing flows associated with entity
func (f *Flow) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Flow.List")
	defer span.End()

	flows, err := flow.List(ctx, params["entity_id"], f.db)
	if err != nil {
		return err
	}

	viewModelFlows := make([]flow.ViewModelFlow, len(flows))
	for i, flow := range flows {
		viewModelFlows[i] = createViewModelFlow(flow)
	}

	return web.Respond(ctx, w, viewModelFlows, http.StatusOK)
}

// Retrieve returns the specified flow from the system.
func (f *Flow) Retrieve(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Flow.Retrieve")
	defer span.End()

	flow, err := flow.Retrieve(ctx, params["flow_id"], f.db)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, createViewModelFlow(flow), http.StatusOK)
}

// Create inserts a new flow into the entity.
func (f *Flow) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Flow.Create")
	defer span.End()

	var nf flow.NewFlow
	if err := web.Decode(r, &nf); err != nil {
		return errors.Wrap(err, "")
	}
	nf.ID = uuid.New().String()
	nf.AccountID = params["account_id"]
	nf.EntityID = params["entity_id"]

	log.Printf("Flow ::::: %+v", nf)

	flow, err := flow.Create(ctx, f.db, nf, time.Now())
	if err != nil {
		return errors.Wrapf(err, "Flow: %+v", &flow)
	}

	return web.Respond(ctx, w, flow, http.StatusCreated)
}

func createViewModelFlow(f flow.Flow) flow.ViewModelFlow {
	return flow.ViewModelFlow{
		ID:          f.ID,
		Name:        f.Name,
		Description: f.Description,
		Expression:  f.Expression,
	}
}

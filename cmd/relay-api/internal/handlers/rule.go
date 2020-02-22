package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/rule"
	"go.opencensus.io/trace"
)

// Rule represents the Rule API method handler set.
type Rule struct {
	db            *sqlx.DB
	authenticator *auth.Authenticator
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// List returns all the existing rules associated with entity
func (i *Rule) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Rule.List")
	defer span.End()

	rules, err := rule.List(ctx, params["entity_id"], i.db)
	if err != nil {
		return err
	}

	viewModelRules := make([]rule.ViewModelRule, len(rules))
	for i, rule := range rules {
		viewModelRules[i] = createViewModelRule(rule)
	}

	return web.Respond(ctx, w, viewModelRules, http.StatusOK)
}

// Create inserts a new rule into the entity.
func (i *Rule) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Rule.Create")
	defer span.End()

	var nr rule.NewRule
	if err := web.Decode(r, &nr); err != nil {
		return errors.Wrap(err, "")
	}
	nr.EntityID = params["entity_id"]

	rule, err := rule.Create(ctx, i.db, nr, time.Now())
	if err != nil {
		return errors.Wrapf(err, "Rule: %+v", &rule)
	}

	return web.Respond(ctx, w, rule, http.StatusCreated)
}

func createViewModelRule(r rule.Rule) rule.ViewModelRule {
	return rule.ViewModelRule{
		ID:         r.ID,
		EntityID:   r.EntityID,
		Expression: r.Expression,
	}
}

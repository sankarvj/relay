package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/ruler"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"go.opencensus.io/trace"
)

type Dashboard struct {
	db            *sqlx.DB
	rPool         *redis.Pool
	authenticator *auth.Authenticator
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

func (d *Dashboard) Overview(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Dashboard.Overview")
	defer span.End()

	exp := r.URL.Query().Get("exp")

	e, err := entity.RetrieveFixedEntity(ctx, d.db, params["account_id"], params["team_id"], "tasks")
	if err == nil {
		return err
	}

	fields := e.FieldsIgnoreError()
	conditionFields := make([]graphdb.Field, 0)
	whoFieldsMap := entity.WhoFieldsMap(fields)

	filter := overdue(whoFieldsMap, entity.WhoDueBy)

	for _, f := range fields {
		if condition, ok := filter.Conditions[f.Key]; ok {
			conditionFields = append(conditionFields, makeGraphField(&f, condition.Term, condition.Expression))
		}
	}

	_, countMap, err := NewSegmenter(exp).
		AddCount().
		filterWrapper(ctx, params["account_id"], e.ID, fields, map[string]interface{}{}, d.db, d.rPool)
	if err != nil {
		return err
	}

	log.Println("countMap---> ", countMap)

	return web.Respond(ctx, w, "", http.StatusOK)
}

func overdue(fields map[string]entity.Field, who string) *ruler.Filter {
	f := fields[who]
	filter := &ruler.Filter{
		Conditions: map[string]ruler.Condition{},
	}
	condition := ruler.Condition{
		Expression: lexertoken.GTSign,
		Term:       util.GetMilliSecondsFloat(time.Now()),
		DataType:   ruler.DType(f.DataType),
	}
	filter.Conditions[f.Key] = condition

	return filter
}

func (d *Dashboard) Count(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	conditionFields := make([]graphdb.Field, 0)
	accountID := params["account_id"]
	e, err := entity.RetrieveFixedEntity(ctx, d.db, params["account_id"], params["team_id"], "tasks")
	if err != nil {
		return errors.Wrapf(err, "entity retieve error")
	}

	fields, err := e.FilteredFields()
	if err != nil {
		return errors.Wrapf(err, "fields retieve error")
	}

	for _, f := range fields {
		if f.Who == entity.WhoStatus {
			conditionFields = append(conditionFields, relatable(f))
		}
	}

	//{Operator:in Key:uuid-00-contacts DataType:S Value:6eb4f58e-8327-4ccc-a262-22ad809e76cb}
	gSegment := graphdb.BuildGNode(accountID, e.ID, false).MakeBaseGNode("", conditionFields)

	result, err := graphdb.GetCount(d.rPool, gSegment, true, true)
	if err != nil {
		return errors.Wrapf(err, "get count")
	}
	log.Printf("result ---%+v", result)

	return web.Respond(ctx, w, "", http.StatusOK)
}

func relatable(f entity.Field) graphdb.Field {
	return graphdb.Field{
		Key:      f.Key,
		Value:    []interface{}{""},
		DataType: graphdb.TypeReference,
		RefID:    f.RefID,
		Field:    &graphdb.Field{},
	}
}

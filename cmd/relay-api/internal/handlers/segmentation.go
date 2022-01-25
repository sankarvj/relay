package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	rg "github.com/redislabs/redisgraph-go"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/reference"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

// Segmentation represents the Segmentation API method handler set.
type Segmentation struct {
	db            *sqlx.DB
	rPool         *redis.Pool
	authenticator *auth.Authenticator
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

func (s *Segmentation) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	var filterBo FilterBody
	if err := web.Decode(r, &filterBo); err != nil {
		return errors.Wrap(err, "")
	}

	nf := flow.NewFlow{
		ID:         uuid.New().String(),
		AccountID:  params["account_id"],
		EntityID:   params["entity_id"],
		Mode:       flow.FlowModeSegment,
		Type:       flow.FlowTypeUnknown,
		Condition:  flow.FlowConditionNil,
		Expression: filterBo.Exp,
		Name:       filterBo.Name,
	}

	f, err := flow.Create(ctx, s.db, nf, time.Now())
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, createViewModelFlow(f, []node.ViewModelNode{}), http.StatusCreated)
}

func filterWrapper(ctx context.Context, accountID, entityID string, fields []entity.Field, exp string, state int, page int, db *sqlx.DB, rp *redis.Pool) ([]ViewModelItem, error) {
	items, err := filterItems(ctx, accountID, entityID, exp, state, page, db, rp)
	if err != nil {
		return nil, err
	}
	viewModelItems := itemResponse(items)
	reference.UpdateReferenceFields(ctx, accountID, entityID, fields, items, map[string]interface{}{}, db, job.NewJabEngine())
	return viewModelItems, nil
}

func filterItems(ctx context.Context, accountID, entityID string, exp string, state int, page int, db *sqlx.DB, rp *redis.Pool) ([]item.Item, error) {

	result, err := segment(ctx, accountID, entityID, exp, page, db, rp)
	if err != nil {
		return nil, err
	}
	items, err := itemsResp(ctx, db, accountID, result)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func segment(ctx context.Context, accountID, entityID string, exp string, page int, db *sqlx.DB, rp *redis.Pool) (*rg.QueryResult, error) {
	conditionFields := make([]graphdb.Field, 0)
	filter := job.NewJabEngine().RunExpGrapher(ctx, db, rp, accountID, exp)
	if filter != nil {
		e, err := entity.Retrieve(ctx, accountID, entityID, db)
		if err != nil {
			return nil, err
		}

		fields, err := e.FilteredFields()
		if err != nil {
			return nil, err
		}

		for _, f := range fields {
			if condition, ok := filter.Conditions[f.Key]; ok {
				conditionFields = append(conditionFields, makeGraphField(&f, condition.Term, condition.Expression))
			}
		}
	}

	//{Operator:in Key:uuid-00-contacts DataType:S Value:6eb4f58e-8327-4ccc-a262-22ad809e76cb}
	gSegment := graphdb.BuildGNode(accountID, entityID, false).MakeBaseGNode("", conditionFields)
	return graphdb.GetResult(rp, gSegment, page)
}

type FilterBody struct {
	Name string `json:"name"`
	Exp  string `json:"exp"`
}

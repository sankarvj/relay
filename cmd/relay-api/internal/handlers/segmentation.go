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
	rg "github.com/redislabs/redisgraph-go"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/reference"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"go.opencensus.io/trace"
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

func (s *Segmentation) Segment(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Segmentation.List")
	defer span.End()

	var seg Segment
	if err := web.Decode(r, &seg); err != nil {
		return errors.Wrap(err, "")
	}

	e, err := entity.Retrieve(ctx, params["account_id"], params["entity_id"], s.db)
	if err != nil {
		return err
	}

	fields, err := e.FilteredFields()
	if err != nil {
		return err
	}

	conditionFields := make([]graphdb.Field, 0)
	for _, f := range fields {
		if condition, ok := seg.Conditions[f.Key]; ok {
			conditionFields = append(conditionFields, makeGraphField(&f, condition.Term, condition.Expression))
		}
	}

	gSegment := graphdb.BuildGNode(params["account_id"], params["entity_id"], false).MakeBaseGNode("", conditionFields)

	result, err := graphdb.GetResult(s.rPool, gSegment)
	if err != nil {
		return errors.Wrap(err, "")
	}

	items, err := itemsResp(ctx, s.db, params["account_id"], e, result)
	if err != nil {
		return err
	}

	fields, viewModelItems := itemResponse(e, items)
	reference.UpdateReferenceFields(ctx, params["account_id"], params["entity_id"], fields, items, map[string]interface{}{}, s.db, job.NewJabEngine())

	response := struct {
		Items    []ViewModelItem        `json:"items"`
		Category int                    `json:"category"`
		Fields   []entity.Field         `json:"fields"`
		Entity   entity.ViewModelEntity `json:"entity"`
	}{
		Items:    viewModelItems,
		Category: e.Category,
		Fields:   fields,
		Entity:   createViewModelEntity(e),
	}

	return web.Respond(ctx, w, response, http.StatusOK)

}

func filterItems(ctx context.Context, accountID, entityID, exp, viewID string, state int, db *sqlx.DB, rp *redis.Pool) (interface{}, error) {
	e, err := entity.Retrieve(ctx, accountID, entityID, db)
	if err != nil {
		return nil, err
	}

	var items []item.Item
	if viewID == "" && exp == "" {
		var err error
		items, err = item.ListFilterByState(ctx, e.ID, state, db)
		if err != nil {
			return nil, err
		}
	} else {
		if exp == "" {
			fl, err := flow.Retrieve(ctx, viewID, db)
			if err != nil {
				return nil, err
			}
			exp = fl.Expression
		}
		result, err := segment(ctx, accountID, e.ID, exp, db, rp)
		if err != nil {
			return nil, err
		}
		items, err = itemsResp(ctx, db, accountID, e, result)
		if err != nil {
			return nil, err
		}
	}

	fields, viewModelItems := itemResponse(e, items)
	reference.UpdateReferenceFields(ctx, accountID, entityID, fields, items, map[string]interface{}{}, db, job.NewJabEngine())

	response := struct {
		Items    []ViewModelItem        `json:"items"`
		Category int                    `json:"category"`
		Fields   []entity.Field         `json:"fields"`
		Entity   entity.ViewModelEntity `json:"entity"`
	}{
		Items:    viewModelItems,
		Category: e.Category,
		Fields:   fields,
		Entity:   createViewModelEntity(e),
	}
	return response, nil
}

func segment(ctx context.Context, accountID, entityID string, exp string, db *sqlx.DB, rp *redis.Pool) (*rg.QueryResult, error) {
	conditionFields := make([]graphdb.Field, 0)
	log.Printf("exp -- %+v", exp)
	filter := job.NewJabEngine().RunExpGrapher(ctx, db, rp, accountID, exp)
	log.Printf("filter -- %+v", filter)
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

	log.Printf("conditionFields -- %+v", conditionFields)
	//{Operator:in Key:uuid-00-contacts DataType:S Value:6eb4f58e-8327-4ccc-a262-22ad809e76cb}
	gSegment := graphdb.BuildGNode(accountID, entityID, false).MakeBaseGNode("", conditionFields)
	return graphdb.GetResult(rp, gSegment)
}

func makeGraphField(f *entity.Field, value interface{}, expression string) graphdb.Field {
	if f.IsReference() {
		return graphdb.Field{
			Key:       f.Key,
			Value:     []interface{}{""},
			DataType:  graphdb.TypeReference,
			RefID:     f.RefID,
			IsReverse: false,
			Field: &graphdb.Field{
				Expression: graphdb.Operator(expression),
				Key:        "id",
				DataType:   graphdb.TypeString,
				Value:      value,
			},
		}
	} else if f.IsList() {
		return graphdb.Field{
			Key:      f.Key,
			Value:    []interface{}{value},
			DataType: graphdb.DType(f.DataType),
			Field: &graphdb.Field{
				Expression: graphdb.Operator(expression),
				Key:        "element",
				DataType:   graphdb.DType(f.Field.DataType),
			},
		}
	} else {
		return graphdb.Field{
			Expression: graphdb.Operator(expression),
			Key:        f.Key,
			DataType:   graphdb.DType(f.DataType),
			Value:      value,
		}
	}
}

type Segment struct {
	Conditions map[string]Condition `json:"conditions"` // key is the field key
}

type Condition struct {
	Term       interface{} `json:"term"`
	Expression string      `json:"expression"`
}

func itemsResp(ctx context.Context, db *sqlx.DB, accountID string, e entity.Entity, result *rg.QueryResult) ([]item.Item, error) {
	items, err := item.BulkRetrieveItems(ctx, accountID, itemIDs(result), db)
	if err != nil {
		return []item.Item{}, err
	}

	return items, nil
}

func itemIDs(result *rg.QueryResult) []interface{} {
	itemIds := make([]interface{}, 0)
	for result.Next() { // Next returns true until the iterator is depleted.
		// Get the current Record.
		r := result.Record()

		// Entries in the Record can be accessed by index or key.
		record := util.ConvertInterfaceToMap(util.ConvertInterfaceToMap(r.GetByIndex(0))["Properties"])
		itemIds = append(itemIds, record["id"])
	}
	return itemIds
}

type FilterBody struct {
	Name string `json:"name"`
	Exp  string `json:"exp"`
}

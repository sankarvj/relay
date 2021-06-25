package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/reference"
	"go.opencensus.io/trace"
)

// Segmentation represents the Segmentation API method handler set.
type Segmentation struct {
	db            *sqlx.DB
	rPool         *redis.Pool
	authenticator *auth.Authenticator
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
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

	log.Printf("conditionFields %+v", conditionFields)

	gSegment := graphdb.BuildGNode(params["account_id"], params["entity_id"], false).MakeBaseGNode("", conditionFields)

	log.Printf("gSegment %+v", gSegment)

	result, err := graphdb.GetResult(s.rPool, gSegment)
	if err != nil {
		return errors.Wrap(err, "")
	}

	itemIds := make([]interface{}, 0)
	for result.Next() { // Next returns true until the iterator is depleted.
		// Get the current Record.
		r := result.Record()

		// Entries in the Record can be accessed by index or key.
		record := util.ConvertInterfaceToMap(util.ConvertInterfaceToMap(r.GetByIndex(0))["Properties"])
		itemIds = append(itemIds, record["id"])
	}

	items, err := item.BulkRetrieveItems(ctx, params["account_id"], itemIds, s.db)
	if err != nil {
		return err
	}

	viewModelItems := make([]item.ViewModelItem, len(items))
	for i, item := range items {
		viewModelItems[i] = createViewModelItem(item)
	}

	reference.UpdateReferenceFields(ctx, params["account_id"], fields, viewModelItems, map[string]interface{}{}, s.db, job.NewJabEngine())

	response := struct {
		Items    []item.ViewModelItem   `json:"items"`
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

func makeGraphField(f *entity.Field, value interface{}, expression string) graphdb.Field {
	if f.IsReference() {
		return graphdb.Field{
			Key:       f.Key,
			Value:     []interface{}{""},
			DataType:  graphdb.TypeReference,
			RefID:     f.RefID,
			IsReverse: false,
			Field: &graphdb.Field{
				Expression: expression,
				Key:        "id",
				DataType:   graphdb.TypeString,
				Value:      value,
			},
		}
	} else if f.IsList() {
		return graphdb.Field{
			Key:      f.Key,
			Value:    value,
			DataType: graphdb.DType(f.DataType),
			Field: &graphdb.Field{
				Expression: expression,
				Key:        "element",
				DataType:   graphdb.DType(f.Field.DataType),
			},
		}
	} else {
		return graphdb.Field{
			Expression: expression,
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

package handlers

import (
	"context"
	"fmt"
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

	expression, tokens := makeExpression(filterBo.Queries)

	nf := flow.NewFlow{
		ID:         uuid.New().String(),
		AccountID:  params["account_id"],
		EntityID:   params["entity_id"],
		Mode:       flow.FlowModeSegment,
		Type:       flow.FlowTypeUnknown,
		Condition:  flow.FlowConditionNil,
		Expression: expression,
		Tokens:     tokens,
		Name:       filterBo.Name,
	}

	f, err := flow.Create(ctx, s.db, nf, time.Now())
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, createViewModelFlow(f, []node.ViewModelNode{}), http.StatusCreated)
}

func (s Segmenter) filterWrapper(ctx context.Context, accountID, entityID string, fields []entity.Field, state int, db *sqlx.DB, rp *redis.Pool) ([]ViewModelItem, map[string]int, error) {
	itemResultBody, err := s.filterItems(ctx, accountID, entityID, state, db, rp)
	if err != nil {
		return nil, nil, err
	}
	viewModelItems := itemResponse(itemResultBody.Items)
	reference.UpdateReferenceFields(ctx, accountID, entityID, fields, itemResultBody.Items, map[string]interface{}{}, db, job.NewJabEngine())
	return viewModelItems, itemResultBody.TotalCount, nil
}

func (s Segmenter) filterItems(ctx context.Context, accountID, entityID string, state int, db *sqlx.DB, rp *redis.Pool) (*ItemResultBody, error) {

	segmentResult, countResult, err := s.segment(ctx, accountID, entityID, db, rp)
	if err != nil {
		return nil, err
	}
	items, err := itemsResp(ctx, db, accountID, segmentResult)
	if err != nil {
		return nil, err
	}
	var totalCount map[string]int
	if s.CountEnabled() {
		totalCount = counts(countResult)
	}

	return &ItemResultBody{Items: items, TotalCount: totalCount}, nil
}

func (s Segmenter) segment(ctx context.Context, accountID, entityID string, db *sqlx.DB, rp *redis.Pool) (*rg.QueryResult, *rg.QueryResult, error) {
	log.Printf("segmenter %+v\n ----> ", s)
	conditionFields := make([]graphdb.Field, 0)

	filter := job.NewJabEngine().RunExpGrapher(ctx, db, rp, accountID, s.exp)
	log.Printf("filter ----> %+v\n", filter)
	if filter != nil {
		e, err := entity.Retrieve(ctx, accountID, entityID, db)
		if err != nil {
			return nil, nil, err
		}

		fields, err := e.FilteredFields()
		if err != nil {
			return nil, nil, err
		}

		for _, f := range fields {
			if condition, ok := filter.Conditions[f.Key]; ok {
				conditionFields = append(conditionFields, makeGraphField(&f, condition.Term, condition.Expression))
			}
		}
	}

	//{Operator:in Key:uuid-00-contacts DataType:S Value:6eb4f58e-8327-4ccc-a262-22ad809e76cb}
	gSegment := graphdb.BuildGNode(accountID, entityID, false).MakeBaseGNode("", conditionFields)

	log.Printf("gSegment--> %+v\n", gSegment)
	return listWithCountAsync(rp, gSegment, s.page, s.sortby, s.direction, s.CountEnabled())
}

func listWithCountAsync(rp *redis.Pool, gSegment graphdb.GraphNode, page int, sortby, direction string, doCount bool) (*rg.QueryResult, *rg.QueryResult, error) {
	//going async way
	loopCount := 1
	type dbResult struct {
		result *rg.QueryResult
		_type  string
	}

	resc, errc := make(chan dbResult), make(chan error)
	go func(rPool *redis.Pool, gSegment graphdb.GraphNode, pageNo int, sortBy, direction string) {
		result, err := graphdb.GetResult(rp, gSegment, page, sortby, direction)
		if err != nil {
			errc <- err
			return
		}
		resc <- dbResult{result: result, _type: "segment"}
	}(rp, gSegment, page, sortby, direction)
	if doCount {
		loopCount = 2
		go func(rPool *redis.Pool, gCount graphdb.GraphNode) {
			result, err := graphdb.GetCount(rPool, gCount, false, false)
			if err != nil {
				errc <- err
				return
			}
			resc <- dbResult{result: result, _type: "count"}
		}(rp, gSegment)
	}

	var err error
	var segmentResult *rg.QueryResult
	var countResult *rg.QueryResult

	for i := 0; i < loopCount; i++ {
		select {
		case dbResult := <-resc:
			switch dbResult._type {
			case "segment":
				segmentResult = dbResult.result
			case "count":
				countResult = dbResult.result
			}
		case err := <-errc:
			fmt.Println(err)
		}
	}
	return segmentResult, countResult, err
}

type ItemResultBody struct {
	Items      []item.Item    `json:"items"`
	TotalCount map[string]int `json:"total_count"`
}

type FilterBody struct {
	Name    string       `json:"name"`
	Queries []node.Query `json:"queries"`
}

type Segmenter struct {
	exp       string
	sortby    string
	direction string
	page      int
	doCount   bool
}

func NewSegmenter(exp string) *Segmenter {
	return &Segmenter{exp: exp}
}

func (s *Segmenter) AddSortLogic(sortby, direction string) *Segmenter {
	s.sortby = sortby
	s.direction = direction
	return s
}

func (s *Segmenter) AddPage(page int) *Segmenter {
	s.page = page
	return s
}

func (s *Segmenter) AddCount() *Segmenter {
	s.doCount = true
	return s
}

func (s *Segmenter) CountEnabled() bool {
	return s.doCount && s.page == 0
}

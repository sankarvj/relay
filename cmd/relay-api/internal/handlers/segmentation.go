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
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/reference"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

// Segmentation represents the Segmentation API method handler set.
type Segmentation struct {
	db            *sqlx.DB
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

func (s Segmenter) filterWrapper(ctx context.Context, accountID, entityID string, fields []entity.Field, sourceMap map[string]interface{}, db *sqlx.DB, sdb *database.SecDB) ([]ViewModelItem, map[string]int, error) {
	itemResultBody, err := s.filterItems(ctx, accountID, entityID, db, sdb)
	if err != nil {
		return nil, nil, err
	}
	//adding users
	userIDs := make(map[string]bool, 0)
	for _, item := range itemResultBody.Items {
		userIDs[*item.UserID] = true
	}
	uMap, _ := userMap(ctx, userIDs, db)
	viewModelItems := itemResponse(itemResultBody.Items, uMap)
	reference.UpdateReferenceFields(ctx, accountID, entityID, fields, itemResultBody.Items, sourceMap, db, job.NewJabEngine())
	return viewModelItems, itemResultBody.TotalCount, nil
}

func (s Segmenter) filterItems(ctx context.Context, accountID, entityID string, db *sqlx.DB, sdb *database.SecDB) (*ItemResultBody, error) {

	segmentResult, countResult, err := s.segment(ctx, accountID, entityID, db, sdb)
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

func (s Segmenter) segment(ctx context.Context, accountID, entityID string, db *sqlx.DB, sdb *database.SecDB) (*rg.QueryResult, *rg.QueryResult, error) {
	conditionFields, err := makeConditionsFromExp(ctx, accountID, entityID, s.exp, db, sdb)
	if err != nil {
		return nil, nil, err
	}
	log.Println("conditionFields --- ", conditionFields)
	if s.source != nil {
		conditionFields = append(conditionFields, *s.source)
	}
	//{Operator:in Key:uuid-00-contacts DataType:S Value:6eb4f58e-8327-4ccc-a262-22ad809e76cb}
	gSegment := graphdb.BuildGNode(accountID, entityID, false).MakeBaseGNode("", conditionFields)
	gSegment.UseReturnNode = s.useReturn

	return listWithCountAsync(sdb.GraphPool(), gSegment, s.page, s.sortby, s.direction, s.CountEnabled())
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
			result, err := graphdb.GetCount(rPool, gCount, false)
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

func NewEmptySegmenter() *Segmenter {
	return &Segmenter{}
}

func NewSegmenter(exp string) *Segmenter {
	return &Segmenter{exp: exp}
}

func (s *Segmenter) AddSortLogic(sortby, direction string) *Segmenter {
	s.sortby = sortby
	s.direction = direction
	return s
}

func (s *Segmenter) AddExp(exp string) *Segmenter {
	s.exp = exp
	return s
}

func (s *Segmenter) AddPage(page int) *Segmenter {
	s.page = page
	return s
}

func (s *Segmenter) _useReturn() *Segmenter {
	s.useReturn = true
	return s
}

func (s *Segmenter) AddSourceCondition(sourceEntityID, sourceItemID string) *Segmenter {
	s.source = &graphdb.Field{
		Value:     []interface{}{""},
		DataType:  graphdb.TypeReference,
		RefID:     sourceEntityID,
		IsReverse: false,
		Field: &graphdb.Field{
			Expression: graphdb.Operator("eq"),
			Key:        "id",
			DataType:   graphdb.TypeString,
			Value:      sourceItemID,
		},
	}
	return s
}

func (s *Segmenter) AddSourceIDCondition(ids []string) *Segmenter {
	s.source = &graphdb.Field{
		Expression: "in", //adding IN instead of giving the ID in the MakeBaseGNode
		Key:        "id",
		DataType:   graphdb.TypeWist,
		Value:      ids,
	}
	return s
}

func (s *Segmenter) AddCount() *Segmenter {
	s.doCount = true
	return s
}

func (s *Segmenter) CountEnabled() bool {
	return s.doCount && s.page == 0
}

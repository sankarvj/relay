package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"gitlab.com/vjsideprojects/relay/internal/account"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/job"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/database/dbservice"
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
	conditionFields, err := makeConditionsFromExp(ctx, accountID, entityID, s.exp, db, sdb)
	if err != nil {
		return nil, nil, err
	}
	if !auth.God(ctx) {
		conditionFields = append(conditionFields, publicRecordsOnly())
	}

	if s.source != nil {
		conditionFields = append(conditionFields, *s.source)
	}
	useDB := account.UseDB(ctx, db, accountID)
	items, totalCount, err := dbservice.NewDBservice(useDB, db, sdb).Result(ctx, accountID, entityID, s.sortby, s.direction, s.page, s.doCount, s.useReturn, conditionFields)
	if err != nil {
		return nil, nil, err
	}
	//adding users
	userIDs := make(map[string]bool, 0)
	for _, item := range items {
		userIDs[*item.UserID] = true
	}
	uMap, _ := userMap(ctx, accountID, userIDs, db)
	viewModelItems := itemResponse(items, uMap, fields)
	reference.UpdateReferenceFields(ctx, accountID, entityID, fields, items, sourceMap, db, sdb, job.NewJabEngine())
	return viewModelItems, totalCount, nil
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

func (s *Segmenter) AddSourceRefCondition(key, sourceEntityID, sourceItemID string) *Segmenter {
	s.source = &graphdb.Field{
		Value:     []interface{}{""},
		DataType:  graphdb.TypeReference,
		RefID:     sourceEntityID,
		IsReverse: false,
		Key:       key,
		Field: &graphdb.Field{
			Expression: graphdb.Operator("in"),
			Key:        "id",
			DataType:   graphdb.TypeString,
			Value:      sourceItemID,
		},
	}
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

func (s *Segmenter) DoCount(count bool) *Segmenter {
	s.doCount = count
	return s
}

func (s *Segmenter) CountEnabled() bool {
	return s.doCount && s.page == 0
}

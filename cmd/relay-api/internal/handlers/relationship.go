package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/relationship"
	"go.opencensus.io/trace"
)

// Relationship represents the Relationship API method handler set.
type Relationship struct {
	db            *sqlx.DB
	sdb           *database.SecDB
	authenticator *auth.Authenticator
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// List returns all the existing relationships associated with entity_id
func (rs *Relationship) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Relationship.List")
	defer span.End()

	relationships, err := relationship.List(ctx, rs.db, params["account_id"], params["team_id"], params["entity_id"])
	if err != nil {
		return errors.Wrap(err, "selecting relationships for the entity id")
	}

	return web.Respond(ctx, w, relationships, http.StatusOK)
}

func (rs *Relationship) ChildItems(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Connections.List")
	defer span.End()
	sourceEntityID := params["entity_id"]
	sourceItemID := params["item_id"]
	accountID := params["account_id"]
	relationshipID := params["relationship_id"]
	ls := r.URL.Query().Get("ls")
	exp := r.URL.Query().Get("exp")
	page := util.ConvertStrToInt(r.URL.Query().Get("page"))

	relation, err := relationship.Retrieve(ctx, accountID, relationshipID, rs.db)
	if err != nil {
		return err
	}

	relatedEntityID := relation.SrcEntityID
	if relatedEntityID == sourceEntityID {
		relatedEntityID = relation.DstEntityID
	}

	e, err := entity.Retrieve(ctx, accountID, relatedEntityID, rs.db)
	if err != nil {
		return err
	}

	//There are three ways to fetch the child ids
	// 1. Fetch child item ids by querying the connections table.
	// 2. Fetch child item ids by querying the graph db. tick
	// 3. Fetch child item ids by querying the genie_id (formerly parent_item_id)
	viewModelItems, fields, countMap, err := fetchChildItems(ctx, accountID, sourceEntityID, sourceItemID, exp, page, relation, e, rs.db, rs.sdb)
	if err != nil {
		return err
	}

	piper := Piper{Viable: true}
	if ls == entity.MetaRenderPipe && page == 0 {
		piper.sourceEntityID, piper.sourceItemID = sourceEntityID, sourceItemID
		err := pipeKanban(ctx, accountID, e, &piper, rs.db)
		if err != nil {
			return err
		}
		piper.Viable = true
		piper.LS = entity.MetaRenderPipe

		piper.Items = make(map[string][]ViewModelItem, 0)
		for _, vmi := range viewModelItems {
			if vmi.StageID != nil {
				if _, ok := piper.Items[*vmi.StageID]; !ok {
					piper.Items[*vmi.StageID] = make([]ViewModelItem, 0)
				}
				piper.Items[*vmi.StageID] = append(piper.Items[*vmi.StageID], vmi)
			}
		}
	}

	domFields := make([]entity.Field, 0)
	for _, f := range fields {
		if f.DomType != entity.DomNotApplicable {
			domFields = append(domFields, f)
		}
	}

	response := struct {
		Items    []ViewModelItem        `json:"items"`
		Category int                    `json:"category"`
		Fields   []entity.Field         `json:"fields"`
		Entity   entity.ViewModelEntity `json:"entity"`
		Piper    Piper                  `json:"piper"`
		CountMap map[string]int         `json:"count_map"`
	}{
		Items:    viewModelItems,
		Category: e.Category,
		Fields:   domFields,
		Entity:   createViewModelEntity(e),
		Piper:    piper,
		CountMap: countMap,
	}

	return web.Respond(ctx, w, response, http.StatusOK)
}

func fetchChildItems(ctx context.Context, accountID, sourceEntityID, sourceItemID string, exp string, page int, relation relationship.Relationship, e entity.Entity, db *sqlx.DB, sdb *database.SecDB) ([]ViewModelItem, []entity.Field, map[string]int, error) {
	sourceMap := make(map[string]interface{}, 0)
	sourceMap[sourceEntityID] = sourceItemID

	fields := e.FieldsIgnoreError()
	var err error
	var viewModelItems []ViewModelItem
	var countMap map[string]int
	if relation.FieldID == relationship.FieldAssociationKey { //explicit
		viewModelItems, countMap, err = NewSegmenter(exp).AddPage(page).
			AddCount().
			AddSourceCondition(sourceEntityID, sourceItemID).
			filterWrapper(ctx, accountID, e.ID, fields, sourceMap, db, sdb)
	} else { // implicit straight. tasks are the child of deals because task has a deal field
		if isFieldKeyExist(relation.FieldID, entity.FieldsMap(fields)) {
			newExp := fmt.Sprintf("{{%s.%s}} in {%s}", e.ID, relation.FieldID, sourceItemID)
			exp = util.AddExpression(exp, newExp)
			viewModelItems, countMap, err = NewSegmenter(exp).AddPage(page).
				AddCount().
				filterWrapper(ctx, accountID, e.ID, fields, sourceMap, db, sdb)
		} else { // implicit reverse. contacts are the child of deals because deals has a contact field
			var it item.Item
			it, err = item.Retrieve(ctx, sourceEntityID, sourceItemID, db)
			if err != nil {
				return nil, nil, nil, err
			}
			ids := it.Fields()[relation.FieldID]
			if ids != nil {
				viewModelItems, countMap, err = NewSegmenter(exp).AddPage(page).
					AddCount().
					AddSourceIDCondition(util.ConvertSliceTypeRev(ids.([]interface{}))).
					filterWrapper(ctx, accountID, e.ID, fields, sourceMap, db, sdb)
			}
		}
	}

	return viewModelItems, fields, countMap, err
}

func isFieldKeyExist(fieldID string, fields map[string]interface{}) bool {
	if _, ok := fields[fieldID]; ok {
		return true
	}
	return false
}

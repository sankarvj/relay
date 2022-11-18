package handlers

import (
	"context"
	"fmt"
	"log"
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

	relationships, err := relationship.List(ctx, rs.db, params["account_id"], params["team_id"], params["entity_id"], auth.God(ctx))
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

	e, err := entity.Retrieve(ctx, accountID, relatedEntityID, rs.db, rs.sdb)
	if err != nil {
		return err
	}

	fields := e.EasyFieldsByRole(ctx)

	piper := Piper{
		Items: make(map[string][]ViewModelItem, 0),
	}
	var viewModelItems []ViewModelItem
	var countMap map[string]int

	if ls == entity.MetaRenderPipe && page == 0 {
		piper.Items = make(map[string][]ViewModelItem, 0)
		err := loadPiperNodes(ctx, accountID, sourceEntityID, sourceItemID, &piper, rs.db, rs.sdb)
		if err != nil {
			return err
		}

		for _, node := range piper.Nodes {
			piper.NodeKey = e.NodeField().Key
			newExp := fmt.Sprintf("{{%s.%s}} eq {%s}", e.ID, e.NodeField().Key, node.ID)
			finalExp := util.AddExpression(exp, newExp)
			vitems, _, err := fetchChildItems(ctx, accountID, sourceEntityID, sourceItemID, finalExp, 0, relation, e, fields, false, rs.db, rs.sdb)
			if err != nil {
				return err
			}
			piper.Items[node.ID] = vitems
		}
	} else {
		//There are three ways to fetch the child ids
		// 1. Fetch child item ids by querying the connections table.
		// 2. Fetch child item ids by querying the graph db. tick
		// 3. Fetch child item ids by querying the genie_id (formerly parent_item_id)
		viewModelItems, countMap, err = fetchChildItems(ctx, accountID, sourceEntityID, sourceItemID, exp, page, relation, e, fields, true, rs.db, rs.sdb)
		if err != nil {
			return err
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
		Fields:   fields, // this fields passed by reference in various places and choices are populated
		Entity:   createViewModelEntity(e),
		Piper:    piper,
		CountMap: countMap,
	}

	return web.Respond(ctx, w, response, http.StatusOK)
}

func fetchChildItems(ctx context.Context, accountID, sourceEntityID, sourceItemID string, exp string, page int, relation relationship.Relationship, e entity.Entity, fields []entity.Field, count bool, db *sqlx.DB, sdb *database.SecDB) ([]ViewModelItem, map[string]int, error) {
	sourceMap := make(map[string]interface{}, 0)
	sourceMap[sourceEntityID] = sourceItemID

	var err error
	var viewModelItems []ViewModelItem
	var countMap map[string]int
	if relation.FieldID == relationship.FieldAssociationKey { //explicit

		viewModelItems, countMap, err = NewSegmenter(exp).AddPage(page).
			DoCount(count).
			AddSourceCondition(sourceEntityID, sourceItemID).
			filterWrapper(ctx, accountID, e.ID, fields, sourceMap, db, sdb)
		log.Println("sourceEntityID-- ", sourceEntityID)
		log.Println("sourceItemID-- ", sourceItemID)
		log.Println("sourceMap-- ", sourceMap)
		log.Println("viewModelItems-- ", viewModelItems)
	} else { // implicit straight. tasks are the child of deals because task has a deal field
		if isFieldKeyExist(relation.FieldID, entity.KeyValueMap(fields)) {
			// newExp := fmt.Sprintf("{{%s.%s}} in {%s}", e.ID, relation.FieldID, sourceItemID)
			// exp = util.AddExpression(exp, newExp)
			viewModelItems, countMap, err = NewSegmenter(exp).AddPage(page).
				DoCount(count).
				AddSourceCondition(sourceEntityID, sourceItemID).
				filterWrapper(ctx, accountID, e.ID, fields, sourceMap, db, sdb)
		} else { // implicit reverse. contacts are the child of deals because deals has a contact field
			var it item.Item
			it, err = item.Retrieve(ctx, accountID, sourceEntityID, sourceItemID, db)
			if err != nil {
				return nil, nil, err
			}
			ids := it.Fields()[relation.FieldID]
			if ids != nil {
				viewModelItems, countMap, err = NewSegmenter(exp).AddPage(page).
					DoCount(count).
					AddSourceIDCondition(util.ConvertSliceTypeRev(ids.([]interface{}))).
					filterWrapper(ctx, accountID, e.ID, fields, sourceMap, db, sdb)
			}
		}
	}

	return viewModelItems, countMap, err
}

func isFieldKeyExist(fieldID string, fields map[string]interface{}) bool {
	if _, ok := fields[fieldID]; ok {
		return true
	}
	return false
}

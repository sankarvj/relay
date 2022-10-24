package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"go.opencensus.io/trace"
)

// Search returns the items for the given term & key
func (i *Item) Search(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Item.Search")
	defer span.End()

	accountID, entityID, _ := takeAEI(ctx, params, i.db)
	key := r.URL.Query().Get("k")
	term := r.URL.Query().Get("t")
	fi := r.URL.Query().Get("fi")

	e, err := entity.Retrieve(ctx, accountID, entityID, i.db)
	if err != nil {
		return err
	}

	var choices []entity.Choice
	// Its a fixed wrapper entity. Call the respective items
	if e.Category == entity.CategoryFlow { // temp flow handler
		// fi is the entityID here
		choices, err = LikeSearchFlows(ctx, accountID, fi, term, i.db)
		if err != nil {
			return err
		}
	} else if e.Category == entity.CategoryNode { // temp flow handler
		choices, err = LikeSearchNodes(ctx, accountID, []string{fi}, term, i.db)
		if err != nil {
			return err
		}
	} else if e.Category == entity.CategoryChildUnit || term == "" {
		items, err := item.SearchByKey(ctx, e.ID, key, term, i.db)
		if err != nil {
			return err
		}
		whoMap := e.WhoFields()
		layouts := e.Layouts()
		choices = choiceResponse(key, items, whoMap, layouts)
	} else {
		exp := fmt.Sprintf("{{%s.%s}} lk {%s}", e.ID, key, term)
		if e.Field(key).DataType == entity.TypeList {
			choices, err = likeSearchElements(ctx, accountID, e.ID, exp, i.db, i.sdb)
		} else {
			choices, err = likeSearchRefItems(ctx, accountID, e.ID, exp, key, e.WhoFields(), e.Layouts(), i.db, i.sdb)
		}
		if err != nil {
			return err
		}
	}

	log.Printf("choices %+v", choices)

	return web.Respond(ctx, w, choices, http.StatusOK)
}

func likeSearchRefItems(ctx context.Context, accountID, entityID, exp, key string, whoMap, layoutMap map[string]string, db *sqlx.DB, sdb *database.SecDB) ([]entity.Choice, error) {
	result, _, err := NewSegmenter(exp).
		segment(ctx, accountID, entityID, db, sdb)
	if err != nil {
		return nil, err
	}
	items, err := itemsResp(ctx, db, accountID, result)

	if err != nil {
		return nil, err
	}
	return choiceResponse(key, items, whoMap, layoutMap), nil
}

func likeSearchElements(ctx context.Context, accountID, entityID, exp string, db *sqlx.DB, sdb *database.SecDB) ([]entity.Choice, error) {
	duplicateReducer := make(map[string]interface{}, 0)
	choices := make([]entity.Choice, 0)
	result, _, err := NewSegmenter(exp).
		_useReturn().
		segment(ctx, accountID, entityID, db, sdb)
	if err != nil {
		return nil, err
	}
	elements := itemElements(result)

	for _, e := range elements {
		duplicateReducer[e.(string)] = e
	}

	for k, v := range duplicateReducer {
		choice := entity.Choice{
			ID:           k,
			DisplayValue: v,
		}
		choices = append(choices, choice)
	}
	return choices, nil
}

func LikeSearchFlows(ctx context.Context, accountID, entityID, term string, db *sqlx.DB) ([]entity.Choice, error) {
	choices := make([]entity.Choice, 0)

	flows, err := flow.SearchByKey(ctx, accountID, entityID, term, db)
	if err != nil {
		return nil, err
	}
	for _, flow := range flows {
		choice := entity.Choice{
			ID:           flow.ID,
			DisplayValue: flow.Name,
		}
		choices = append(choices, choice)
	}

	return choices, nil
}

func LikeSearchNodes(ctx context.Context, accountID string, flowIDs []string, term string, db *sqlx.DB) ([]entity.Choice, error) {
	choices := make([]entity.Choice, 0)
	nodes, err := node.Stages(ctx, accountID, flowIDs, term, db)
	if err != nil {
		return nil, err
	}
	for _, node := range nodes {
		choice := entity.Choice{
			ID:           node.ID,
			DisplayValue: node.Name,
		}
		choices = append(choices, choice)
	}

	return choices, nil
}

func choiceResponse(key string, items []item.Item, whoMap map[string]string, layoutMap map[string]string) []entity.Choice {
	choices := make([]entity.Choice, 0)
	log.Printf("items %+v", items)
	log.Printf("whoMap %+v", whoMap)
	log.Printf("key %+v", key)
	for _, item := range items {
		log.Printf("item %+v", item)
		//display
		var displayV interface{}
		if key != "" {
			displayV = item.Fields()[key]
		}
		// if key is not passed. Choose the title layout
		if displayV == nil {
			if keyOfDis, ok := layoutMap[entity.MetaLayoutTitle]; ok {
				if title, ok := item.Fields()[keyOfDis]; ok {
					displayV = title.(string)
				}
			}
		}

		//avatar
		var avatar string
		if ava, ok := whoMap[entity.WhoAvatar]; ok {
			if aval, ok := item.Fields()[ava]; ok {
				avatar = aval.(string)
			}
		}

		choice := entity.Choice{
			ID:           item.ID,
			DisplayValue: displayV,
			Avatar:       avatar,
		}
		choices = append(choices, choice)
	}
	return choices
}

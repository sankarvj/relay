package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
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
	filterID := r.URL.Query().Get("fi")

	e, err := entity.Retrieve(ctx, accountID, entityID, i.db)
	if err != nil {
		return err
	}
	choices := make([]entity.Choice, 0)
	// Its a fixed wrapper entity. Call the respective items
	if e.Category == entity.CategoryFlow { // temp flow handler
		flows, err := flow.SearchByKey(ctx, accountID, entityID, key, term, i.db)
		if err != nil {
			return err
		}
		for _, flow := range flows {
			choice := entity.Choice{
				ID:           flow.ID,
				DisplayValue: flow.Name,
			}
			choices = append(choices, choice)
		}
	} else if e.Category == entity.CategoryNode { // temp flow handler
		//here filterID is the flowID...
		nodes, err := node.SearchByKey(ctx, accountID, filterID, key, term, i.db)
		if err != nil {
			return err
		}
		for _, node := range nodes {
			choice := entity.Choice{
				ID:           node.ID,
				DisplayValue: node.Name,
			}
			choices = append(choices, choice)
		}
	} else if e.Category == entity.CategoryChildUnit || term == "" {
		items, err := item.SearchByKey(ctx, e.ID, key, term, i.db)
		if err != nil {
			return err
		}
		whoMap := e.WhoFields()
		choices = choiceResponse(key, items, whoMap)
	} else {
		exp := fmt.Sprintf("{{%s.%s}} lk {%s}", e.ID, key, term)
		if e.Field(key).DataType == entity.TypeList {
			choices, err = likeSearchElements(ctx, accountID, e.ID, exp, i.db, i.rPool)
		} else {
			choices, err = likeSearchRefItems(ctx, accountID, e.ID, exp, key, e.WhoFields(), i.db, i.rPool)
		}
		if err != nil {
			return err
		}
	}

	return web.Respond(ctx, w, choices, http.StatusOK)
}

func likeSearchRefItems(ctx context.Context, accountID, entityID, exp, key string, whoMap map[string]string, db *sqlx.DB, rPool *redis.Pool) ([]entity.Choice, error) {
	result, _, err := NewSegmenter(exp).
		segment(ctx, accountID, entityID, db, rPool)
	if err != nil {
		return nil, err
	}
	items, err := itemsResp(ctx, db, accountID, result)
	if err != nil {
		return nil, err
	}
	return choiceResponse(key, items, whoMap), nil
}

func likeSearchElements(ctx context.Context, accountID, entityID, exp string, db *sqlx.DB, rPool *redis.Pool) ([]entity.Choice, error) {
	duplicateReducer := make(map[string]interface{}, 0)
	choices := make([]entity.Choice, 0)
	result, _, err := NewSegmenter(exp).
		_useReturn().
		segment(ctx, accountID, entityID, db, rPool)
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

func choiceResponse(key string, items []item.Item, whoMap map[string]string) []entity.Choice {
	choices := make([]entity.Choice, 0)
	for _, item := range items {
		//display
		displayV := item.Fields()[key]
		if displayV == nil {
			displayV = item.Name
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

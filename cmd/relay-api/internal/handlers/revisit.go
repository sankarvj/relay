package handlers

import (
	"context"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

/**

All the functions in this file needs revisit.....


***/

func loadItemsWithGroupByLogic(ctx context.Context, accountID string, e entity.Entity, exp, sortby, direction, groupby string, page int, piper *Piper, db *sqlx.DB, sdb *database.SecDB) error {
	fields := e.EasyFieldsByRole(ctx)
	piper.Items = make(map[string][]ViewModelItem, 0)
	piper.Tokens = make(map[string]string, 0)
	piper.Exps = make(map[string]string, 0)
	piper.CountMap = make(map[string]map[string]int, 0)
	choicers, err := groupBy(ctx, groupby, e, db, sdb)
	if err != nil {
		return err
	}
	for _, choicer := range choicers {
		newExp := fmt.Sprintf("{{%s.%s}} in {%s}", e.ID, groupby, choicer.ID)
		if choicer.ID == "" { // get none values
			newExp = fmt.Sprintf("{{%s.%s}} !in {%s}", e.ID, groupby, choicer.Value)
		}
		finalExp := util.AddExpression(exp, newExp)
		vitems, countMap, err := NewSegmenter(finalExp).
			AddPage(page).
			AddSortLogic(sortby, direction).
			AddCount().
			filterWrapper(ctx, accountID, e.ID, fields, map[string]interface{}{}, db, sdb)
		if err != nil {
			return err
		}
		piper.CountMap[choicer.ID] = countMap
		piper.Items[choicer.ID] = vitems
		piper.Tokens[choicer.ID] = choicer.Name
		piper.Exps[choicer.ID] = newExp
	}
	return nil
}

func loadPiperNodes(ctx context.Context, accountID, sourceEntityID, sourceItemID string, piper *Piper, db *sqlx.DB, sdb *database.SecDB) error {
	piper.sourceEntityID, piper.sourceItemID = sourceEntityID, sourceItemID
	se, err := entity.Retrieve(ctx, accountID, sourceEntityID, db, sdb)
	if err != nil {
		return err
	}
	it, err := item.Retrieve(ctx, accountID, sourceEntityID, sourceItemID, db)
	if err != nil {
		return err
	}
	f := se.FlowField()
	if f != nil {
		flowIDs := it.Fields()[f.Key]
		if flowIDs != nil {
			flowID := flowIDs.([]interface{})[0]
			viewModelNodes, err := nodeStages(ctx, accountID, flowID.(string), db)
			if err != nil {
				return err
			}
			piper.Nodes = viewModelNodes
		}
	}
	return nil
}

func loadNodes(ctx context.Context, accountID, sourceEntityID, sourceItemID string, db *sqlx.DB, sdb *database.SecDB) ([]node.NodeActor, error) {
	se, err := entity.Retrieve(ctx, accountID, sourceEntityID, db, sdb)
	if err != nil {
		return nil, err
	}
	it, err := item.Retrieve(ctx, accountID, sourceEntityID, sourceItemID, db)
	if err != nil {
		return nil, err
	}
	f := se.FlowField()
	if f != nil {
		flowIDs := it.Fields()[f.Key]
		if flowIDs != nil {
			flowIDs := flowIDs.([]interface{})
			if len(flowIDs) > 0 {
				flowID := flowIDs[0]
				if flowID != nil && flowID != "" {
					nodes, err := node.NodeActorsList(ctx, accountID, flowID.(string), db)
					if err != nil {
						log.Printf("***> unexpected error occurred when retriving reference items for nodes inside updating choices error: %v.\n continuing...", err)
					}
					return nodes, err
				}
			}

		}
	}
	return []node.NodeActor{}, err
}

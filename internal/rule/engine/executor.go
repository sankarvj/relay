package engine

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func (ruleResult *RuleResult) executePosCase(ctx context.Context, db *sqlx.DB, n node.Node, eng *Engine) error {
	log.Println("executePosCase ActorID ---> ", n.ActorID)
	ruleResult.Executed = true
	var err error
	executionResponse := map[string]interface{}{}

	switch n.Type {
	case node.Push:
		err = eng.executeData(ctx, db, n)
	case node.Modify:
		err = eng.executeData(ctx, db, n)
	case node.Hook:
		var result map[string]interface{}
		result, err = executeHook(ctx, db, n)
		executionResponse[node.GlobalEntityData] = result
	case node.Email:
		err = eng.executeEmail(ctx, db, n)
	case node.Decision:
		err = nil
		executionResponse[node.GlobalEntityResult] = true
	case node.Delay:
		err = executeDelay(ctx, db, n)
	case node.Stage:
		err = nil
	}

	ruleResult.Response = map[string]interface{}{node.GlobalEntity: executionResponse}
	return err
}

func (ruleResult *RuleResult) executeNegCase(ctx context.Context, db *sqlx.DB, n node.Node) error {
	ruleResult.Executed = false
	executionResponse := map[string]interface{}{}
	switch n.Type {
	case node.Decision:
		ruleResult.Executed = true //because the decision is considered as executed even it is in false condition
		executionResponse[node.GlobalEntityResult] = false
	}
	ruleResult.Response = map[string]interface{}{node.GlobalEntity: executionResponse}
	return nil
}

func valueAdd(ctx context.Context, db *sqlx.DB, accountID, entityID, itemID string) ([]entity.Field, error) {
	e, err := entity.Retrieve(ctx, accountID, entityID, db)
	if err != nil {
		return []entity.Field{}, err
	}
	if itemID != "" {
		i, err := item.Retrieve(ctx, entityID, itemID, db)
		if err != nil {
			return []entity.Field{}, err
		}
		return e.ValueAdd(i.Fields()), nil
	}

	return []entity.Field{}, nil
}

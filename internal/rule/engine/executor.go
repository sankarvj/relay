package engine

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func (ruleResult *RuleResult) executePosCase(ctx context.Context, db *sqlx.DB, n node.Node) error {
	log.Println("executePosCase ActorID ---> ", n.ActorID)
	ruleResult.Executed = true
	var err error
	executionResponse := map[string]interface{}{}

	switch n.Type {
	case node.Push:
		err = executeData(ctx, db, n)
	case node.Modify:
		err = executeData(ctx, db, n)
	case node.Hook:
		var result map[string]interface{}
		result, err = executeHook(ctx, db, n)
		executionResponse[node.GlobalEntityData] = result
	case node.Email:
		err = executeEmail(ctx, db, n)
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

func fields(ctx context.Context, db *sqlx.DB, accountID, entityID string) ([]entity.Field, error) {
	//Load entity maps If valid entity exists.
	e, err := entity.Retrieve(ctx, accountID, entityID, db)
	if err != nil {
		return []entity.Field{}, err
	}
	return e.Fields()
}

func fillItemFieldValues(ctx context.Context, db *sqlx.DB, entityFields []entity.Field, entityID string, itemIDs ...string) ([]entity.Field, error) {
	for _, itemID := range itemIDs {
		if itemID != "" {
			i, err := item.Retrieve(ctx, entityID, itemID, db)
			if err != nil {
				return nil, err
			}
			entityFields = entity.ValueAddFields(entityFields, i.Fields())
		}
	}

	return entityFields, nil
}

func mergeActualsWithActor(ctx context.Context, db *sqlx.DB, accountID, actorEntityID, actorItemID string) ([]entity.Field, error) {
	entityFields, err := fields(ctx, db, accountID, actorEntityID)
	if err != nil {
		return nil, err
	}

	entityFields, err = fillItemFieldValues(ctx, db, entityFields, actorEntityID, actorItemID)
	if err != nil {
		return nil, err
	}
	return entityFields, nil
}

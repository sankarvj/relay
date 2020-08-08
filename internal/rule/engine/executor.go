package engine

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
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
		executionResponse[GlobalEntityData] = result
	case node.Email:
		err = executeEmail(ctx, db, n)
	case node.Decision:
		err = nil
		executionResponse[GlobalEntityResult] = true
	case node.Delay:
		err = executeDelay(ctx, db, n)
	case node.Stage:
		err = nil
	}

	ruleResult.Response = map[string]interface{}{GlobalEntity: executionResponse}
	return err
}

func (ruleResult *RuleResult) executeNegCase(ctx context.Context, db *sqlx.DB, n node.Node) error {
	log.Println("executeNegCase ActorID ---> ", n.ActorID)
	ruleResult.Executed = false
	executionResponse := map[string]interface{}{}
	switch n.Type {
	case node.Decision:
		ruleResult.Executed = true //because the decision is considered as executed even it is in false condition
		executionResponse[GlobalEntityResult] = false
	}
	ruleResult.Response = map[string]interface{}{GlobalEntity: executionResponse}
	return nil
}

func namedFieldsMap(entityFields []entity.Field) map[string]interface{} {
	params := map[string]interface{}{}
	for _, field := range entityFields {
		params[field.Name] = field.Value
	}
	return params
}

func fields(ctx context.Context, db *sqlx.DB, entityID string) ([]entity.Field, error) {
	//Load entity maps If valid entity exists.
	e, err := entity.Retrieve(ctx, entityID, db)
	if err != nil {
		return []entity.Field{}, err
	}
	return e.AllFields()
}
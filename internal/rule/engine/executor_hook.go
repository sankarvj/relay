package engine

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func executeHook(ctx context.Context, db *sqlx.DB, n node.Node) (map[string]interface{}, error) {
	e, err := entity.Retrieve(ctx, n.AccountID, n.ActorID, db)
	if err != nil {
		return map[string]interface{}{}, err
	}

	result, err := retriveAPIEntityResult(e.FieldsIgnoreError())
	log.Printf("rule.engine.executor_hook: result: %s and err: %v\n", result, err)
	return result, err
}

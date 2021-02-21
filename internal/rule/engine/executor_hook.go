package engine

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func executeHook(ctx context.Context, db *sqlx.DB, n node.Node) (map[string]interface{}, error) {
	entityFields, err := fields(ctx, db, n.AccountID, n.ActorID)
	if err != nil {
		return map[string]interface{}{}, err
	}
	result, err := retriveAPIEntityResult(entityFields)
	log.Println("result :: ", result)
	log.Println("err :: ", err)
	return result, err
}

package engine

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func executePosCase(ctx context.Context, db *sqlx.DB, n node.Node) (map[string]interface{}, error) {
	executionResponse := map[string]interface{}{}
	log.Println("execute ActorID ---> ", n.ActorID)

	e, err := entity.Retrieve(ctx, n.ActorID, db)
	if err != nil {
		return executionResponse, err
	}

	entityFields, err := e.AllFields()
	if err != nil {
		return executionResponse, err
	}

	switch n.Type {
	case node.Push:
		err = executeData(ctx, db, entityFields, n)
	case node.Modify:
		err = executeData(ctx, db, entityFields, n)
	case node.Hook:
		var result map[string]interface{}
		result, err = executeHook(ctx, db, entityFields)
		executionResponse["data"] = result
	case node.Email:
		err = executeEmail(ctx, db, entityFields, n)
	case node.Decision:
		err = nil
	}

	if err != nil {
		executionResponse["result"] = false
	} else {
		executionResponse["result"] = true
	}

	// switch e.Category {
	// case entity.CategoryAPI:
	// 	executeHook(ctx, db, entityFields)
	// case entity.CategoryEmail:
	// 	executeEmail(ctx, db, entityFields, n)
	// case entity.CategoryData:
	// 	executeData(ctx, db, entityFields, n)
	// }

	return map[string]interface{}{GlobalEntity: executionResponse}, err
}

func executeNegCase(ctx context.Context, db *sqlx.DB, n node.Node) (map[string]interface{}, error) {
	executionResponse := map[string]interface{}{}
	switch n.Type {
	case node.Decision:
		executionResponse["result"] = false
	}
	return executionResponse, nil
}

func namedFieldsMap(entityFields []entity.Field) map[string]interface{} {
	params := map[string]interface{}{}
	for _, field := range entityFields {
		params[field.Name] = field.Value
	}
	return params
}

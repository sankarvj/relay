package engine

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func executePosCase(ctx context.Context, db *sqlx.DB, n node.Node) (map[string]interface{}, error) {
	log.Println("executePosCase ActorID ---> ", n.ActorID)
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
		executionResponse["data"] = result
	case node.Email:
		err = executeEmail(ctx, db, n)
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
	log.Println("executeNegCase ActorID ---> ", n.ActorID)
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

func fields(ctx context.Context, db *sqlx.DB, entityID string) ([]entity.Field, error) {
	//Load entity maps If valid entity exists.
	e, err := entity.Retrieve(ctx, entityID, db)
	if err != nil {
		return []entity.Field{}, err
	}
	return e.AllFields()
}

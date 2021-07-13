package engine

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/net"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/ruler"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

func worker(ctx context.Context, db *sqlx.DB, accountID string, expression string, input map[string]interface{}) (interface{}, error) {
	log.Printf("running worker for expression %s : %v", expression, input)
	entityID := ruler.FetchEntityID(expression)
	if entityID == node.GlobalEntity { //global entity stops here.
		return input, nil
	} else if entityID == node.SelfEntity { //self entity stops here
		return buildResultant(node.SelfEntity, input), nil
	}
	//TODO cache entity
	e, err := entity.Retrieve(ctx, accountID, entityID, db)
	if err != nil {
		return map[string]interface{}{}, err
	}
	var result map[string]interface{}
	switch e.Category {
	case entity.CategoryAPI:
		fields, err := e.Fields()
		if err != nil {
			return map[string]interface{}{}, err
		}
		result, err = retriveAPIEntityResult(fields)
	case entity.CategoryData, entity.CategoryTimeSeries:
		if itemID, ok := input[e.ID]; ok {
			// TODO itemID.(string) we are blindly typecasting it to string???
			// TODO cache the object instead of itemID
			// what happens if different data type comes??
			result, err = retriveDataEntityResult(ctx, db, e.ID, itemID.(string))
		}
	}
	if err != nil {
		result = map[string]interface{}{"error": err}
	}
	return Evaluate(expression, buildResultant(e.ID, result)), nil
}

func retriveAPIEntityResult(fields []entity.Field) (map[string]interface{}, error) {
	var result map[string]interface{}
	apiParams, err := populateAPIParams(fields)
	if err != nil {
		return result, err
	}
	err = apiParams.MakeHTTPRequest(&result)
	return result, err
}

func retriveDataEntityResult(ctx context.Context, db *sqlx.DB, entityID, itemID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	item, err := item.Retrieve(ctx, entityID, itemID, db)
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal([]byte(item.Fieldsb), &result); err != nil {
		return result, errors.Wrapf(err, "error while unmarshalling item attributes on retrive with fields %q", item.ID)
	}
	//sets the id as one of the field key to make use of the reference fields
	result["id"] = itemID
	return result, err
}

func populateAPIParams(entityFields []entity.Field) (net.APIParams, error) {
	var webHookEntityItem entity.WebHookEntity
	err := entity.ParseFixedEntity(entityFields, &webHookEntityItem)
	if err != nil {
		return net.APIParams{}, err
	}
	apiParams := net.APIParams{
		Path:    webHookEntityItem.Path,
		Host:    webHookEntityItem.Host,
		Method:  webHookEntityItem.Method,
		Headers: webHookEntityItem.Headers,
	}
	return apiParams, nil
}

func buildResultant(entityID string, result map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		entityID: result,
	}
}

//Evaluate evaluates the expression with the coresponding map
func Evaluate(expression string, response map[string]interface{}) interface{} {
	var realValue interface{}
	elements := strings.Split(expression, ".")
	lenOfElements := len(elements)
	for index, element := range elements {
		if index == (lenOfElements - 1) {
			realValue = response[element]
			break
		}
		if response[element] == nil {
			break
		}
		response = response[element].(map[string]interface{})
	}
	return realValue
}

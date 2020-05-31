package rule

import (
	"context"
	"encoding/json"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/net"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/ruler"
)

func worker(ctx context.Context, db *sqlx.DB, expression string, input map[string]string) map[string]interface{} {
	log.Println("running worker for expression ", expression)
	entityID := ruler.FetchEntityID(expression)
	e, err := entity.Retrieve(ctx, entityID, db)
	if err != nil {
		return map[string]interface{}{"error": err}
	}
	var result map[string]interface{}
	switch e.Category {
	case entity.CategoryAPI:
		fields, err := e.Fields()
		if err != nil {
			return map[string]interface{}{"error": err}
		}
		result, err = retriveAPIEntityResult(fields)
	case entity.CategoryData, entity.CategoryTimeSeries:
		if val, ok := input[e.ID]; ok {
			result, err = retriveDataEntityResult(ctx, db, e.ID, val)
		}
	}
	if err != nil {
		result = map[string]interface{}{"error": err}
	}
	return buildResultant(e.ID, result)
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
	item, err := item.Retrieve(ctx, itemID, db)
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal([]byte(item.Input), &result); err != nil {
		return result, errors.Wrapf(err, "error while unmarshalling item attributes on retrive with fields %q", item.ID)
	}
	return result, err
}

func populateAPIParams(fields []entity.Field) (net.APIParams, error) {
	apiParams := net.APIParams{}
	for _, field := range fields {
		switch field.Name {
		case "path":
			apiParams.Path = field.Value
		case "host":
			apiParams.Host = field.Value
		case "method":
			apiParams.Method = field.Value
		case "headers":
			var headers map[string]string
			if err := json.Unmarshal([]byte(field.Value), &headers); err != nil {
				return apiParams, err
			}
			apiParams.Headers = headers
		}
	}
	return apiParams, nil
}

func buildResultant(entityID string, result map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		entityID: result,
	}
}

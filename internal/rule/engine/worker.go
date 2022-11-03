package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/net"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/ruler"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"gitlab.com/vjsideprojects/relay/internal/user"
)

func worker(ctx context.Context, db *sqlx.DB, sdb *database.SecDB, accountID string, expression string, input map[string]interface{}) (interface{}, error) {
	log.Printf("internal.rule.engine.worker running expression: %s\n", expression)
	entityID := ruler.FetchEntityID(expression)
	if entityID == node.GlobalEntity { //global entity stops here.
		return input, nil
	} else if entityID == node.SelfEntity { //self entity stops here
		return evaluate(ctx, db, sdb, accountID, expression, buildResultant(node.SelfEntity, input)), nil
	} else if entityID == node.MeEntity { //replace with currentuser_id
		currentUserID, err := user.RetrieveCurrentUserID(ctx)
		if err != nil {
			return nil, err
		}
		return memberID(ctx, db, accountID, currentUserID)
	} else if entityID == node.SegmentEntity { // flow expression with enters/leaves segment
		return ruler.FetchItemID(expression), nil
	}
	//TODO cache entity
	e, err := entity.Retrieve(ctx, accountID, entityID, db, sdb)
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

	finalResult := evaluate(ctx, db, sdb, accountID, expression, buildResultant(e.ID, result))
	return finalResult, nil
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
func evaluate(ctx context.Context, db *sqlx.DB, sdb *database.SecDB, accountID, expression string, response map[string]interface{}) interface{} {
	var realValue interface{}
	elements := strings.Split(expression, ".")

	entityID := ""
	fieldKey := ""
	superBug := "" // It is the temp concept used in the invite visitors template to get email-ids from the associated elements
	if len(elements) > 2 {
		entityID = elements[0]
		fieldKey = elements[1]
		superBug = elements[2]
	}

	lenOfElements := len(elements)
	for index, element := range elements {
		if index == (lenOfElements - 1) {
			realValue = response[element]
			break
		}

		if response[element] == nil {
			break
		}

		switch t := response[element].(type) {
		case []interface{}:
			realValue = superBugger(ctx, db, sdb, accountID, entityID, fieldKey, response[element], superBug)
			response[superBug] = realValue
		case map[string]interface{}:
			response = t
		}
	}
	return realValue
}

//Not so useful as of now
func superBugger(ctx context.Context, db *sqlx.DB, sdb *database.SecDB, accountID, entityID, fieldKey string, response interface{}, suberBug string) string {
	output := ""
	switch t := response.(type) {
	case []interface{}:
		e, err := entity.Retrieve(ctx, accountID, entityID, db, sdb)
		if err != nil {
			log.Printf("***> unexpected error occurred on superBugger when retriving entity  - error: %v ", err)
			return ""
		}

		f := e.Field(fieldKey)

		itemIds := make([]interface{}, 0)
		if suberBug == node.EmailEntityData {
			itemIds = append(itemIds, t...)
		}

		if len(itemIds) > 0 {
			e, err = entity.Retrieve(ctx, accountID, f.RefID, db, sdb)
			if err != nil {
				log.Printf("***> unexpected error occurred on superBugger when retriving entity  - error: %v ", err)
				return ""
			}
			items, err := item.BulkRetrieve(ctx, f.RefID, itemIds, db)
			if err != nil {
				log.Printf("***> unexpected error occurred on superBugger when retriving items  - error: %v ", err)
				return ""
			}

			emailFields := e.OnlyEmailFields()
			for _, it := range items {
				for _, ef := range emailFields {
					email := it.Fields()[ef.Key]
					if output == "" {
						output = email.(string)
					} else {
						output = fmt.Sprintf("%s, %s", output, email)
					}
				}

			}
		}

	default:
		output = t.(string)
	}
	return output
}

func memberID(ctx context.Context, db *sqlx.DB, accountID, userID string) (string, error) {
	creator, err := user.RetrieveUser(ctx, db, userID)
	if err != nil {
		return "", err
	}
	if memberID, ok := creator.AccountsB()[accountID]; ok {
		return memberID.(string), nil
	}
	return "", errors.New("member not found")
}

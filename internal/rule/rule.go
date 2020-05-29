package rule

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/net"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/ruler"
	"go.opencensus.io/trace"
)

// List retrieves a list of existing rules for the entity associated from the database.
func List(ctx context.Context, entityID string, db *sqlx.DB) ([]Rule, error) {
	ctx, span := trace.StartSpan(ctx, "internal.rule.List")
	defer span.End()

	rules := []Rule{}
	const q = `SELECT * FROM rules where entity_id = $1`

	if err := db.SelectContext(ctx, &rules, q, entityID); err != nil {
		return nil, errors.Wrap(err, "selecting rules")
	}

	return rules, nil
}

// Create inserts a new item into the database.
func Create(ctx context.Context, db *sqlx.DB, n NewRule, now time.Time) (*Rule, error) {
	ctx, span := trace.StartSpan(ctx, "internal.rule.Create")
	defer span.End()

	actionItems, err := json.Marshal(n.ActionItems)
	if err != nil {
		return nil, errors.Wrap(err, "encode action")
	}

	r := Rule{
		ID:         uuid.New().String(),
		EntityID:   n.EntityID,
		Expression: n.Expression + " <" + string(actionItems) + ">",
		CreatedAt:  now.UTC(),
		UpdatedAt:  now.UTC().Unix(),
	}

	const q = `INSERT INTO rules
		(rule_id, entity_id, expression, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err = db.ExecContext(
		ctx, q,
		r.ID, r.EntityID, r.Expression,
		r.CreatedAt, r.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "inserting rule")
	}

	return &r, nil
}

//RunRuleEngine runs the engine on the expression and emit the action to be taken
func RunRuleEngine(ctx context.Context, db *sqlx.DB, exp string) {
	action := make(chan string)
	work := make(chan ruler.Work)
	go ruler.Run(exp, work, action)
	go startWorker(ctx, db, work)
	for {
		actionItem, ok := <-action
		if !ok {
			fmt.Println("Channel Close 1")
			break
		}
		log.Println("action Item", actionItem)
		//actioner(ctx, db, actionItem)
	}
}

func startWorker(ctx context.Context, db *sqlx.DB, w chan ruler.Work) {
	for {
		work, ok := <-w
		if !ok {
			fmt.Println("Channel Close 2")
			break
		}

		work.Resp <- worker(ctx, db, work.Expression)
	}
}

func worker(ctx context.Context, db *sqlx.DB, expression string) map[string]interface{} {
	key := ruler.FetchRootKey(expression)
	e, fields, err := entity.RetrieveWithFields(ctx, db, key)
	if err != nil {
		return map[string]interface{}{"error": errors.Wrapf(err, "error while retriving entity on response worker %v", key)}
	}

	var result map[string]interface{}
	switch e.Category {
	case entity.CategoryAPI:
		err = updateResultFromAPIEntity(fields, &result)
	case entity.CategoryData:
		result, err = updateResultFromDataEntity(ctx, db, expression, result)
	}
	if err != nil {
		result = map[string]interface{}{"error": err}
	}

	return buildResultant(e.ID, result)
}

func updateResultFromAPIEntity(fields []entity.Field, result *map[string]interface{}) error {
	apiParams, err := populateAPIParams(fields)
	if err != nil {
		return err
	}
	err = apiParams.MakeHTTPRequest(result)
	return err
}

func updateResultFromDataEntity(ctx context.Context, db *sqlx.DB, expression string, result map[string]interface{}) (map[string]interface{}, error) {
	key := ruler.FetchRootKey(expression)
	itemType := ruler.FetchItemType(expression)
	switch itemType {
	case "latest":
		_, fields, err := item.RetrieveLatestItem(ctx, db, key)
		result = map[string]interface{}{
			itemType: fields,
		}
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

func populateAPIParams(fields []entity.Field) (net.APIParams, error) {
	apiParams := net.APIParams{}
	for _, field := range fields {
		switch field.Key {
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

func buildResultant(rootKey string, result map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		rootKey: result,
	}
}

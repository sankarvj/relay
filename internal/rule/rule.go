package rule

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
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

	r := Rule{
		ID:         uuid.New().String(),
		EntityID:   n.EntityID,
		Expression: n.Expression,
		CreatedAt:  now.UTC(),
		UpdatedAt:  now.UTC().Unix(),
	}

	const q = `INSERT INTO rules
		(rule_id, entity_id, expression, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := db.ExecContext(
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
func RunRuleEngine(exp string, db *sqlx.DB) {
	action := make(chan string)
	go ruler.Run(exp, dummypayal, action)

	for {
		act, ok := <-action
		if !ok {
			fmt.Println("Channel Close")
			break
		}
		fmt.Println("Action To Be Taken ", act)
	}
}

func dummypayal(key string) map[string]interface{} {

	// entity, err := entity.Retrieve(ctx, key, e.db)
	// if err != nil {
	// 	return err
	// }

	if key == "a6036fe2-0e77-4fab-a798-a39fcf99815c" {
		return map[string]interface{}{
			"a6036fe2-0e77-4fab-a798-a39fcf99815c": map[string]interface{}{
				"build": map[string]interface{}{
					"artifact": 1,
					"appinfo": map[string]interface{}{
						"version": 2,
					},
				},
			},
		}
	} else {
		return map[string]interface{}{
			"8ac6147e-ad53-4379-8503-806c01500b9b": map[string]interface{}{
				"latest": map[string]interface{}{
					"version": 2,
				},
			},
		}
	}
}

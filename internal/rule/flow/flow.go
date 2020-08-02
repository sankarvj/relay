package flow

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/rule/engine"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"go.opencensus.io/trace"
)

var (
	// ErrNotFound is used when a specific flow is requested but does not exist.
	ErrNotFound = errors.New("Flow not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")
)

// List retrieves a list of existing flows for the entity change.
func List(ctx context.Context, entityID string, db *sqlx.DB) ([]Flow, error) {
	ctx, span := trace.StartSpan(ctx, "internal.rule.flow.List")
	defer span.End()

	flows := []Flow{}
	const q = `SELECT * FROM flows where entity_id = $1`

	if err := db.SelectContext(ctx, &flows, q, entityID); err != nil {
		return nil, errors.Wrap(err, "selecting flows")
	}

	return flows, nil
}

// Create inserts a new item into the database.
func Create(ctx context.Context, db *sqlx.DB, n NewFlow, now time.Time) (Flow, error) {
	ctx, span := trace.StartSpan(ctx, "internal.rule.flow.Create")
	defer span.End()

	f := Flow{
		ID:          uuid.New().String(),
		AccountID:   n.AccountID,
		EntityID:    n.EntityID,
		Name:        n.Name,
		Description: n.Description,
		Expression:  n.Expression,
		Type:        n.Type,
		Condition:   n.Condition,
		Status:      0,
		CreatedAt:   now.UTC(),
		UpdatedAt:   now.UTC().Unix(),
	}

	const q = `INSERT INTO flows
		(flow_id, account_id, entity_id, name, description, expression, type, condition, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := db.ExecContext(
		ctx, q,
		f.ID, f.AccountID, f.EntityID, f.Name, f.Description, f.Expression, f.Type, f.Condition, f.Status,
		f.CreatedAt, f.UpdatedAt,
	)
	if err != nil {
		return Flow{}, errors.Wrap(err, "inserting flow")
	}

	return f, nil
}

// Retrieve gets the specified flow from the database.
func Retrieve(ctx context.Context, id string, db *sqlx.DB) (Flow, error) {
	ctx, span := trace.StartSpan(ctx, "internal.rule.flow.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return Flow{}, ErrInvalidID
	}

	var f Flow
	const q = `SELECT * FROM flows WHERE flow_id = $1`
	if err := db.GetContext(ctx, &f, q, id); err != nil {
		if err == sql.ErrNoRows {
			return Flow{}, ErrNotFound
		}

		return Flow{}, errors.Wrapf(err, "selecting flow %q", id)
	}

	return f, nil
}

// DirtyFlows filters the flows which matches the field name in the rules with the modified fields during the update/insert
// operation of the item on the entity.
func DirtyFlows(ctx context.Context, flows []Flow, oldItemFields, newItemFields map[string]interface{}) []Flow {
	ctx, span := trace.StartSpan(ctx, "internal.rule.flow.LazyFlows")
	defer span.End()

	if len(flows) == 0 {
		return flows
	}

	dirtyFlows := make([]Flow, 0)
	dirtyFields := item.Diff(oldItemFields, newItemFields)
	for key := range dirtyFields {
		for _, flow := range flows {
			if strings.Contains(flow.Expression, key) {
				dirtyFlows = append(dirtyFlows, flow)
			}
		}
	}

	return dirtyFlows
}

// Trigger triggers the inactive flows which are ready to be triggerd based on rules in the flows
func Trigger(ctx context.Context, itemID string, dirtyFlows []Flow, db *sqlx.DB) error {
	aflows, err := ActiveFlows(ctx, ids(dirtyFlows), db)
	if err != nil {
		return err
	}
	activeFlowMap := activeFlowMap(aflows)
	for _, df := range dirtyFlows {
		af := activeFlowMap[df.ID]
		if evaluateExpression(ctx, itemID, df.EntityID, df.Expression, db) { //entry
			err = af.entryFlowTrigger(ctx, db, df.AccountID, df.ID, df.EntityID, itemID, df.Type, allowFlowEntryTrigger(df.Condition))
		} else {
			err = af.exitFlowTrigger(ctx, db, df.AccountID, df.ID, df.EntityID, itemID, df.Type, allowFlowExitTrigger(df.Condition))
		}
		//concat errors in the loop. nil if no error exists
		err = errors.Wrapf(err, "error in entry/exit trigger for flowID %q", df.ID)
	}

	return err
}

//DirectTrigger is when you want to execute the item on a particular node stage.
func DirectTrigger(ctx context.Context, db *sqlx.DB, n node.Node, entityID, itemID string, flowType int) error {
	af, afErr := RetrieveAF(ctx, db, itemID, n.FlowID)
	if afErr != nil && afErr != ErrNotFound {
		return afErr
	}

	an, err := RetrieveAN(ctx, db, n.ID, itemID, n.FlowID)
	if err != ErrNotFound {
		return ErrNodeAlreadyActive
	}

	if evaluateExpression(ctx, itemID, entityID, n.Expression, db) {
		err = upsertAF(ctx, db, af, n, itemID, afErr)
		if err != nil {
			return err
		}
		return an.entryNodeTrigger(ctx, db, n, entityID, itemID, flowType)
	}
	return ErrExpressionConditionFailed
}

func upsertAF(ctx context.Context, db *sqlx.DB, af ActiveFlow, n node.Node, itemID string, err error) error {
	if err == ErrNotFound {
		af.AccountID = n.AccountID
		af.FlowID = n.FlowID
		af.ItemID = itemID
		af.IsActive = true
		af.NodeID = n.ID
		af.Life = 1
		_, err = CreateAF(ctx, db, af)
	} else {
		af.NodeID = n.ID
		err = UpdateAFNode(ctx, db, n.ID, itemID, n.FlowID)
	}
	return err
}

// evaluateExpression evaluates the given expression and returns yes/no. It builds the dynamic variables
// by mapping the changed entity/item and passes those variables to the RunExpEvaluator
func evaluateExpression(ctx context.Context, itemID, entityID, expression string, db *sqlx.DB) bool {
	variables := map[string]interface{}{
		entityID: itemID,
	}
	return engine.RunExpEvaluator(ctx, db, expression, variables)
}

func ids(lazyFlows []Flow) []string {
	ids := make([]string, len(lazyFlows))
	for i, flow := range lazyFlows {
		ids[i] = flow.ID
	}
	return ids
}

func allowFlowEntryTrigger(condition int) bool {
	return condition == FlowConditionBoth || condition == FlowConditionEntry
}

func allowFlowExitTrigger(condition int) bool {
	return condition == FlowConditionBoth || condition == FlowConditionExit
}

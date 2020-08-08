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
	// ErrNotFound occurs when the flow is not found
	ErrNotFound = errors.New("Flow not found")

	// ErrFlowActive occurs when the item asking to enter the active flow again
	ErrFlowActive = errors.New("Cannot enter the flow. Flow is already active for the item")

	// ErrFlowInActive occurs when the item asking to exit from the inactive/new flow
	ErrFlowInActive = errors.New("Cannot exit the flow. Flow is not active for the item")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrInvalidItemEntity occurs when an item's entity is different from the flow entity
	ErrInvalidItemEntity = errors.New("Item cannot be added to the flow. Entity mismatch")

	// ErrInvalidFlowType occurs direct trigger is executed for any flow type but pipeline
	ErrInvalidFlowType = errors.New("This operation is cannot be performed for this flow type")
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
func Trigger(ctx context.Context, db *sqlx.DB, itemID string, dirtyFlows []Flow) error {
	aflows, err := ActiveFlows(ctx, ids(dirtyFlows), db)
	if err != nil {
		return err
	}
	activeFlowMap := activeFlowMap(aflows)
	for _, df := range dirtyFlows {
		af := activeFlowMap[df.ID]
		n := node.RootNode(df.AccountID, df.ID, df.EntityID, itemID, df.Expression).UpdateMeta(df.EntityID, itemID, df.Type)
		if engine.RunExpEvaluator(ctx, db, n.Expression, n.VariablesMap()) { //entry
			if af.stopEntryTriggerFlow(df.Condition) { //skips trigger if already active or of exit condition
				return ErrFlowActive
			}
			err = af.entryFlowTrigger(ctx, db, n)
		} else {
			if af.stopExitTriggerFlow(df.Condition) { //skips trigger if new  or inactive or not allowed.
				return ErrFlowInActive
			}
			err = af.exitFlowTrigger(ctx, db, n)
		}
		//concat errors in the loop. nil if no error exists
		err = errors.Wrapf(err, "error in entry/exit trigger for flowID %q", df.ID)
	}

	return err
}

//DirectTrigger is when you want to execute the item on a particular node stage.
func DirectTrigger(ctx context.Context, db *sqlx.DB, nodeID, itemID string) error {
	//retrival of primary components item,flow,node
	n, err := node.Retrieve(ctx, nodeID, db)
	if err != nil {
		return err
	}
	f, err := Retrieve(ctx, n.FlowID, db)
	if err != nil {
		return err
	}
	i, err := item.Retrieve(ctx, itemID, db)
	if err != nil {
		return err
	}

	//verification of primary components item,flow,node
	if n.Type != node.Stage {
		return node.ErrInvalidNodeType
	}
	if f.Type != FlowTypePipeline {
		return ErrInvalidFlowType
	}
	if i.EntityID != f.EntityID {
		return ErrInvalidItemEntity
	}

	//update meta. very important to update meta before calling exp evalustor
	n.UpdateMeta(i.EntityID, i.ID, f.Type)
	if engine.RunExpEvaluator(ctx, db, n.Expression, n.VariablesMap()) {
		af, err := RetrieveAF(ctx, db, itemID, f.ID)
		if err != nil && err != ErrNotFound {
			return err
		}
		return af.entryFlowTrigger(ctx, db, n)
	}
	return ErrExpressionConditionFailed
}

func ids(lazyFlows []Flow) []string {
	ids := make([]string, len(lazyFlows))
	for i, flow := range lazyFlows {
		ids[i] = flow.ID
	}
	return ids
}

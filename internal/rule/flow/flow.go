package flow

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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

	// ErrInvalidFlowMode occurs direct trigger is executed for any flow mode but pipeline
	ErrInvalidFlowMode = errors.New("This operation is cannot be performed for this flow mode")
)

// List retrieves a list of existing flows for the entity change.
func List(ctx context.Context, entityIDs []string, fm int, db *sqlx.DB) ([]Flow, error) {
	ctx, span := trace.StartSpan(ctx, "internal.rule.flow.List")
	defer span.End()

	flows := []Flow{}
	if len(entityIDs) == 0 {
		return flows, nil
	}
	modes := []int{fm}
	if fm == FlowModeAll {
		modes = []int{FlowModeWorkFlow, FlowModePipeLine}
	}
	q, args, err := sqlx.In(`SELECT * FROM flows where entity_id IN (?) AND mode IN (?);`, entityIDs, modes)
	if err != nil {
		return nil, errors.Wrap(err, "selecting in query")
	}
	q = db.Rebind(q)
	if err := db.SelectContext(ctx, &flows, q, args...); err != nil {
		return nil, errors.Wrap(err, "selecting flows")
	}

	return flows, nil
}

// Create inserts a new item into the database.
func Create(ctx context.Context, db *sqlx.DB, nf NewFlow, now time.Time) (Flow, error) {
	ctx, span := trace.StartSpan(ctx, "internal.rule.flow.Create")
	defer span.End()

	f := Flow{
		ID:          nf.ID,
		AccountID:   nf.AccountID,
		EntityID:    nf.EntityID,
		Name:        nf.Name,
		Description: nf.Description,
		Expression:  nf.Expression,
		Mode:        nf.Mode,
		Type:        nf.Type,
		Condition:   nf.Condition,
		Status:      0,
		CreatedAt:   now.UTC(),
		UpdatedAt:   now.UTC().Unix(),
	}

	const q = `INSERT INTO flows
		(flow_id, account_id, entity_id, name, description, expression, type, mode, condition, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := db.ExecContext(
		ctx, q,
		f.ID, f.AccountID, f.EntityID, f.Name, f.Description, f.Expression, f.Type, f.Mode, f.Condition, f.Status,
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

func SearchByKey(ctx context.Context, accountID, entityID, key, term string, db *sqlx.DB) ([]Flow, error) {
	ctx, span := trace.StartSpan(ctx, "internal.flow.SearchByKey")
	defer span.End()

	flows := []Flow{}
	const q = `SELECT * FROM flows where account_id = $1 AND entity_id = $2`

	if err := db.SelectContext(ctx, &flows, q, accountID, entityID); err != nil {
		return nil, errors.Wrap(err, "searching flows")
	}

	return flows, nil
}

func BulkRetrieve(ctx context.Context, accountID string, ids []interface{}, db *sqlx.DB) ([]Flow, error) {
	ctx, span := trace.StartSpan(ctx, "internal.flow.BulkRetrieve")
	defer span.End()

	flows := []Flow{}
	const q = `SELECT * FROM flows where account_id = $1 AND flow_id = any($2)`

	if err := db.SelectContext(ctx, &flows, q, accountID, pq.Array(ids)); err != nil {
		return flows, errors.Wrap(err, "selecting bulk flows for selected flow ids")
	}

	return flows, nil
}

// DirtyFlows filters the flows which matches the field name in the rules with the modified fields during the update/insert
// operation of the item on the entity.
func DirtyFlows(ctx context.Context, flows []Flow, dirtyFields map[string]interface{}) []Flow {
	ctx, span := trace.StartSpan(ctx, "internal.rule.flow.LazyFlows")
	defer span.End()

	if len(flows) == 0 {
		return flows
	}
	dirtyFlows := make([]Flow, 0)
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
func Trigger(ctx context.Context, db *sqlx.DB, rp *redis.Pool, itemID string, flows []Flow, eng engine.Engine) []error {
	ctx, span := trace.StartSpan(ctx, "internal.rule.flow.Trigger")
	defer span.End()
	triggerErrors := make([]error, 0)

	//TODO what if the matched flows has 1 million records
	aflows, err := ActiveFlows(ctx, ids(flows), db)
	if err != nil {
		return append(triggerErrors, err)
	}
	activeFlowMap := activeFlowMap(aflows)
	for _, f := range flows {
		log.Printf("check expression for flow ->  %s", f.Name)
		af := activeFlowMap[f.ID]
		n := node.RootNode(f.AccountID, f.ID, f.EntityID, itemID, f.Expression).UpdateMeta(f.EntityID, itemID, f.Type).UpdateVariables(f.EntityID, itemID)
		if eng.RunExpEvaluator(ctx, db, rp, n.AccountID, n.Expression, n.VariablesMap()) { //entry
			if af.stopEntryTriggerFlow(f.Condition) { //skip trigger if already active or of exit condition
				err = ErrFlowActive
			} else {
				err = af.entryFlowTrigger(ctx, db, rp, n, eng)
			}
		} else {
			if af.stopExitTriggerFlow(f.Condition) { //skip trigger if new  or inactive or not allowed.
				err = ErrFlowInActive
			} else {
				err = af.exitFlowTrigger(ctx, db, rp, n, eng)
			}
		}
		//concat errors in the loop. nil if no error exists
		err = errors.Wrapf(err, "error in entry/exit trigger for flowID %q", f.ID)
		triggerErrors = append(triggerErrors, err)
	}

	return triggerErrors
}

//DirectTrigger is when you want to execute the item on a particular node stage.
func DirectTrigger(ctx context.Context, db *sqlx.DB, rp *redis.Pool, accountID, flowID, nodeID, entityID, itemID string, eng engine.Engine) error {
	//retrival of primary components item,flow,node
	f, err := Retrieve(ctx, flowID, db)
	if err != nil {
		return err
	}
	n, err := node.Retrieve(ctx, accountID, flowID, nodeID, db)
	if err != nil {
		return err
	}
	i, err := item.Retrieve(ctx, entityID, itemID, db)
	if err != nil {
		return err
	}

	//verification of primary components item,flow,node
	if f.Mode != FlowModePipeLine {
		return ErrInvalidFlowMode
	}
	if n.Type != node.Stage {
		return node.ErrInvalidNodeType
	}
	if i.EntityID != f.EntityID {
		return ErrInvalidItemEntity
	}

	//update meta. very important to update meta before calling exp evaluator
	n.UpdateMeta(i.EntityID, i.ID, f.Type).UpdateVariables(i.EntityID, i.ID)
	if eng.RunExpEvaluator(ctx, db, rp, n.AccountID, n.Expression, n.VariablesMap()) {
		af, err := RetrieveAF(ctx, db, itemID, f.ID)
		if err != nil && err != ErrNotFound { //usually for the first time the af will not exist there
			return err
		}
		return af.entryFlowTrigger(ctx, db, rp, n, eng) // entering the workflow with all the variables loaded in the node
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

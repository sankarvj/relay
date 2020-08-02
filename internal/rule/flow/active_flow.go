package flow

import (
	"context"
	"database/sql"
	"log"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/rule/engine"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"go.opencensus.io/trace"
)

//Errors in the flow
var (
	ErrExpressionConditionFailed = errors.New("Expression failed. Can't move the item to the node")
)

// CreateAF inserts a new item into the active_flows table.
func CreateAF(ctx context.Context, db *sqlx.DB, af ActiveFlow) (ActiveFlow, error) {
	ctx, span := trace.StartSpan(ctx, "internal.flow.activeFlow.Create")
	defer span.End()

	const q = `INSERT INTO active_flows
		(account_id,flow_id, item_id, node_id, is_active, life)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := db.ExecContext(
		ctx, q,
		af.AccountID, af.FlowID, af.ItemID, af.NodeID, af.IsActive, af.Life,
	)
	if err != nil {
		return ActiveFlow{}, errors.Wrap(err, "inserting active_flow")
	}

	return af, nil
}

// UpdateAF modifies the active flow
func UpdateAF(ctx context.Context, db *sqlx.DB, af ActiveFlow) error {
	ctx, span := trace.StartSpan(ctx, "internal.flow.activeFlow.Update")
	defer span.End()

	const q = `UPDATE active_flows SET
		"is_active" = $3,
		"life" = $4
		WHERE item_id = $1 AND flow_id = $2` //should I include account_id in the where clause for sharding?
	_, err := db.ExecContext(ctx, q, af.ItemID, af.FlowID,
		af.IsActive, af.Life,
	)
	if err != nil {
		return errors.Wrap(err, "updating active flow")
	}

	return nil
}

// UpdateAFNode updates the active flow with node_id
func UpdateAFNode(ctx context.Context, db *sqlx.DB, nodeID, itemID, flowID string) error {
	ctx, span := trace.StartSpan(ctx, "internal.flow.activeFlow.Update")
	defer span.End()

	const q = `UPDATE active_flows SET
		"node_id" = $3 
		WHERE item_id = $1 AND flow_id = $2` //should I include account_id in the where clause for sharding?
	_, err := db.ExecContext(ctx, q, itemID, flowID,
		nodeID,
	)
	if err != nil {
		return errors.Wrap(err, "updating active flow with nodeID")
	}

	return nil
}

// RetrieveAF gets the specified active flow from the database.
func RetrieveAF(ctx context.Context, db *sqlx.DB, itemID, flowID string) (ActiveFlow, error) {
	ctx, span := trace.StartSpan(ctx, "internal.flow.activeFlow.Retrive")
	defer span.End()

	if _, err := uuid.Parse(flowID); err != nil {
		return ActiveFlow{}, ErrInvalidID
	}

	var af ActiveFlow
	const q = `SELECT * FROM active_flows WHERE item_id = $1 AND flow_id = $2`
	if err := db.GetContext(ctx, &af, q, itemID, flowID); err != nil {
		if err == sql.ErrNoRows {
			return ActiveFlow{}, ErrNotFound
		}

		return ActiveFlow{}, errors.Wrapf(err, "selecting active flow %q", flowID)
	}

	return af, nil
}

// ActiveFlows get the active flows entries for the dirty flow ids if exists
func ActiveFlows(ctx context.Context, flowIDs []string, db *sqlx.DB) ([]ActiveFlow, error) {
	ctx, span := trace.StartSpan(ctx, "internal.flow.activeFlow.activeFlows")
	defer span.End()

	activeFlows := []ActiveFlow{}
	const q = `SELECT * FROM active_flows where flow_id = any($1)`

	if err := db.SelectContext(ctx, &activeFlows, q, pq.Array(flowIDs)); err != nil {
		return activeFlows, errors.Wrap(err, "selecting active flows")
	}

	return activeFlows, nil
}

//activeFlowMap maps flowID with the activeFlow
func activeFlowMap(activeFlows []ActiveFlow) map[string]ActiveFlow {
	activeFlowMap := map[string]ActiveFlow{}
	for _, aflow := range activeFlows {
		activeFlowMap[aflow.FlowID] = aflow
	}
	return activeFlowMap
}

func (af ActiveFlow) entryFlowTrigger(ctx context.Context, db *sqlx.DB, accountID, flowID, entityID, itemID string, flowType int, allowEntryTrigger bool) error {
	var err error
	if af.IsActive { //skips trigger if already active or of exit condition
		return nil
	}
	af.FlowID = flowID
	af.ItemID = itemID
	af.AccountID = accountID
	af.NodeID = engine.NullID
	af.IsActive = true
	af.Life = af.Life + 1
	if af.Life == 1 {
		_, err = CreateAF(ctx, db, af)
	} else {
		err = UpdateAF(ctx, db, af)
	}
	if err == nil && allowEntryTrigger {
		// this should be called in the jobQ
		err = startJobFlow(ctx, db, flowID, entityID, itemID, flowType)
	}

	return err
}

func (af ActiveFlow) exitFlowTrigger(ctx context.Context, db *sqlx.DB, accountID, flowID, entityID, itemID string, flowType int, allowExitTrigger bool) error {
	if af.Life == 0 || !af.IsActive { //skips trigger if new  or inactive.
		return nil
	}
	af.IsActive = false
	err := UpdateAF(ctx, db, af)

	if err == nil && allowExitTrigger {
		err = startJobFlow(ctx, db, flowID, entityID, itemID, flowType)
	}
	return err
}

func updateVarJSON(existingVars map[string]interface{}, engineRes map[string]interface{}) string {
	return node.VariablesJSON(updateMap(existingVars, engineRes))
}

func updateMap(existingVars map[string]interface{}, newVars map[string]interface{}) map[string]interface{} {
	for key, exitingVal := range existingVars {
		if _, ok := newVars[key]; !ok { //if existing key present in newVars then keep the newVars value.
			newVars[key] = exitingVal
		} else {
			// for the global entity, dive in to the innerMap (inside xyz).
			// we should update the content inside global entities and not just replace it with new values.
			if key == engine.GlobalEntity || key == engine.GlobalEntityData {
				exitingGlobalMap := existingVars[key].(map[string]interface{})
				newGlobalMap := newVars[key].(map[string]interface{})
				newVars[key] = updateMap(exitingGlobalMap, newGlobalMap)
			}
		}
	}
	return newVars
}

func logFlowEvent(ctx context.Context, db *sqlx.DB, n node.Node) {
	log.Printf(">>>>>>>>>>>>>>>>        The Item Has Entered The Segment Flow %v", n.FlowID)
}

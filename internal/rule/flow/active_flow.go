package flow

import (
	"context"
	"database/sql"
	"log"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"go.opencensus.io/trace"
)

//Errors in the flow
var (
	ErrExpressionConditionFailed = errors.New("Expression failed. Can't move the item to the node")
)

// CreateAF inserts a new item into the active_flows table.
func CreateAF(ctx context.Context, db *sqlx.DB, af ActiveFlow) (ActiveFlow, error) {
	ctx, span := trace.StartSpan(ctx, "internal.rule.flow.activeFlow.Create")
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
	ctx, span := trace.StartSpan(ctx, "internal.rule.flow.activeFlow.Update")
	defer span.End()

	const q = `UPDATE active_flows SET
		"node_id" = $3, 
		"is_active" = $4,
		"life" = $5
		WHERE item_id = $1 AND flow_id = $2` //should I include account_id in the where clause for sharding?
	_, err := db.ExecContext(ctx, q, af.ItemID, af.FlowID,
		af.NodeID, af.IsActive, af.Life,
	)
	if err != nil {
		return errors.Wrap(err, "updating active flow")
	}

	return nil
}

// RetrieveAF gets the specified active flow from the database.
func RetrieveAF(ctx context.Context, db *sqlx.DB, itemID, flowID string) (ActiveFlow, error) {
	ctx, span := trace.StartSpan(ctx, "internal.rule.flow.activeFlow.Retrive")
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
// TODO pagination. This is called from the UI
func ActiveFlows(ctx context.Context, flowIDs []string, db *sqlx.DB) ([]ActiveFlow, error) {
	ctx, span := trace.StartSpan(ctx, "internal.rule.flow.activeFlow.activeFlows")
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

func (af ActiveFlow) entryFlowTrigger(ctx context.Context, db *sqlx.DB, rp *redis.Pool, n *node.Node) error {
	_, span := trace.StartSpan(ctx, "internal.rule.flow.Trigger.entryFlowTrigger")
	defer span.End()
	log.Printf("triggering entryflow")
	if err := af.enableAF(ctx, db, n.AccountID, n.FlowID, n.ID, n.Meta.ItemID); err != nil {
		return err
	}

	return startJobFlow(ctx, db, rp, n)
}

func (af ActiveFlow) exitFlowTrigger(ctx context.Context, db *sqlx.DB, rp *redis.Pool, n *node.Node) error {
	_, span := trace.StartSpan(ctx, "internal.rule.flow.Trigger.exitFlowTrigger")
	defer span.End()
	log.Printf("triggering exitflow")
	if err := af.disableAF(ctx, db); err != nil {
		return err
	}
	return startJobFlow(ctx, db, rp, n)
}

func (af ActiveFlow) stopEntryTriggerFlow(condition int) bool {
	return (condition != FlowConditionBoth && condition != FlowConditionEntry) || af.IsActive
}

func (af ActiveFlow) stopExitTriggerFlow(condition int) bool {
	return (condition != FlowConditionExit && condition != FlowConditionBoth) || af.Life == 0 || !af.IsActive
}

func (af ActiveFlow) enableAF(ctx context.Context, db *sqlx.DB, accountID, flowID, nodeID, itemID string) error {
	if af.Life == 0 {
		af.AccountID = accountID
		af.FlowID = flowID
		af.ItemID = itemID
		af.NodeID = nodeID
		af.IsActive = true
		af.Life = 1
		_, err := CreateAF(ctx, db, af)
		return err
	}
	if nodeID != node.Root { // is this works? not setting the nodeID if the node is a Root
		af.NodeID = nodeID
	}
	af.IsActive = true
	af.Life = af.Life + 1
	return UpdateAF(ctx, db, af)
}

func (af ActiveFlow) disableAF(ctx context.Context, db *sqlx.DB) error {
	af.IsActive = false
	return UpdateAF(ctx, db, af)
}

func updateVarJSON(existingVars map[string]interface{}, engineRes map[string]interface{}) string {
	return node.VariablesJSON(node.UpdateNodeVars(existingVars, engineRes))
}

func upsertAF(ctx context.Context, db *sqlx.DB, accountID, flowID, nodeID, itemID string) error {
	af, err := RetrieveAF(ctx, db, itemID, flowID)
	if err != nil && err != ErrNotFound {
		return err
	}
	return af.enableAF(ctx, db, accountID, flowID, nodeID, itemID)
}

func logFlowEvent(ctx context.Context, db *sqlx.DB, n node.Node) {
	log.Printf("the job started for flow %s", n.FlowID)
}

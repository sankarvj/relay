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

//Err vars from node
var (
	ErrNodeAlreadyActive = errors.New("Node is already active. Can't execute it again")
	ErrCannotExecuteNode = errors.New("Node not executed due to expression condition")
)

// CreateAN inserts a new item into the active_nodes table.
func CreateAN(ctx context.Context, db *sqlx.DB, an ActiveNode) (ActiveNode, error) {
	ctx, span := trace.StartSpan(ctx, "internal.rule.flow.activeNode.Create")
	defer span.End()

	const q = `INSERT INTO active_nodes
		(account_id,flow_id, item_id, node_id, is_active, life)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := db.ExecContext(
		ctx, q,
		an.AccountID, an.FlowID, an.ItemID, an.NodeID, an.IsActive, an.Life,
	)
	if err != nil {
		return ActiveNode{}, errors.Wrap(err, "inserting active_node")
	}

	return an, nil
}

// UpdateAN modifies the active node
func UpdateAN(ctx context.Context, db *sqlx.DB, an ActiveNode) error {
	ctx, span := trace.StartSpan(ctx, "internal.rule.flow.activeNode.Update")
	defer span.End()

	const q = `UPDATE active_nodes SET
		"is_active" = $4,
		"life" = $5
		WHERE item_id = $1 AND flow_id = $2 AND node_id = $3` //should I include account_id in the where clause for sharding?
	_, err := db.ExecContext(ctx, q, an.ItemID, an.FlowID, an.NodeID,
		an.IsActive, an.Life,
	)
	if err != nil {
		return errors.Wrap(err, "updating active flow")
	}

	return nil
}

// RetrieveAN gets the specified node from the database.
func RetrieveAN(ctx context.Context, db *sqlx.DB, nodeID, itemID, flowID string) (ActiveNode, error) {
	ctx, span := trace.StartSpan(ctx, "internal.rule.flow.activeNode.Retrive")
	defer span.End()

	if _, err := uuid.Parse(nodeID); err != nil {
		return ActiveNode{}, ErrInvalidID
	}

	var an ActiveNode
	const q = `SELECT * FROM active_nodes WHERE node_id = $1 AND item_id = $2 AND flow_id = $3`
	if err := db.GetContext(ctx, &an, q, nodeID, itemID, flowID); err != nil {
		if err == sql.ErrNoRows {
			return ActiveNode{}, ErrNotFound
		}

		return ActiveNode{}, errors.Wrapf(err, "selecting active node %q", nodeID)
	}

	return an, nil
}

// ActiveNodes get the active nodes entries for the given flows
func ActiveNodes(ctx context.Context, flowIDs []string, db *sqlx.DB) ([]ActiveNode, error) {
	ctx, span := trace.StartSpan(ctx, "internal.rule.flow.activeNode.ActiveNodes")
	defer span.End()

	activeNodes := []ActiveNode{}
	const q = `SELECT * FROM active_nodes where flow_id = any($1)`

	if err := db.SelectContext(ctx, &activeNodes, q, pq.Array(flowIDs)); err != nil {
		return activeNodes, errors.Wrap(err, "selecting active nodes")
	}

	return activeNodes, nil
}

func startJobFlow(ctx context.Context, db *sqlx.DB, n *node.Node) error {
	// call this in job Q
	return nextRun(ctx, db, *n, map[string]interface{}{})
}

func nextRun(ctx context.Context, db *sqlx.DB, n node.Node, parentResponseMap map[string]interface{}) error {
	err := upsertActives(ctx, db, n)
	if err != nil {
		//TODO push this to DL queue
		return err
	}
	nodes, err := node.List(ctx, n.FlowID, db)
	if err != nil {
		//TODO push this to DL queue
		return err
	}
	//if multiple child nodes exists then who will take the job?
	//if the parentNode is a decision node than the result of engine.RunRuleEngine should say result:true/result:false
	//if the parentNode is a hook node than the the result of engine.RunRuleEngine should pass the API response inside the variables
	//if the parentNode is a push/modify/email node than the result of engine.RunRuleEngine should say result:true/result:false
	childNodes := node.ChildNodes(n.ID, node.BranceNodeMap(nodes))
	updatedVars := updateVarJSON(n.VariablesMap(), parentResponseMap)
	for _, childNode := range childNodes {
		childNode.Meta = n.Meta //passing root node meta
		childNode.Variables = updatedVars
		if childNode.Type == node.Stage { //stage nodes should not execute automatically. Always needs a manual intervention
			continue
		}
		runJob(ctx, db, childNode)
	}
	return nil
}

func runJob(ctx context.Context, db *sqlx.DB, n node.Node) error {
	ruleResult, err := engine.RunRuleEngine(ctx, db, n)
	if err != nil {
		//TODO push this to DL queue
		return err
	}
	if !ruleResult.Executed {
		return ErrCannotExecuteNode
	}
	return nextRun(ctx, db, n, ruleResult.Response)
}

func upsertAN(ctx context.Context, db *sqlx.DB, accountID, flowID, nodeID, itemID string) (bool, error) {
	an, err := RetrieveAN(ctx, db, nodeID, itemID, flowID)
	if err != nil && err != ErrNotFound {
		return false, err
	}
	if err == ErrNotFound {
		an.AccountID = accountID
		an.FlowID = flowID
		an.ItemID = itemID
		an.IsActive = true
		an.NodeID = nodeID
		an.Life = 1
		_, err = CreateAN(ctx, db, an)
	} else {
		an.IsActive = true
		an.Life = an.Life + 1
		err = UpdateAN(ctx, db, an)
	}
	return an.Life > 1, err
}

func upsertActives(ctx context.Context, db *sqlx.DB, n node.Node) error {
	log.Printf(">>>>>>>>>>>>>>>>   The Item Has Entered The Node Flow Node ID: %v -- Entity ID: %v -- Item ID: %v -- Flow ID: %v", n.ID, n.Meta.EntityID, n.Meta.ItemID, n.FlowID)
	if n.ID == node.Root { // add the flow entry event
		logFlowEvent(ctx, db, n)
	}

	if err := logNodeEvent(ctx, db, n); err != nil { // add the node entry event
		return err
	}

	if !n.IsRootNode() && n.Meta.FlowType != FlowTypePipeline || (n.Meta.FlowType == FlowTypePipeline && n.IsStageNode()) {
		if _, err := upsertAN(ctx, db, n.AccountID, n.FlowID, n.ID, n.Meta.ItemID); err != nil {
			return err
		}
	}

	return nil
}

func logNodeEvent(ctx context.Context, db *sqlx.DB, n node.Node) error {
	return nil
}

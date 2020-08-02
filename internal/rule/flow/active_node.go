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
	ctx, span := trace.StartSpan(ctx, "internal.flow.activeNode.Create")
	defer span.End()

	const q = `INSERT INTO active_nodes
		(account_id,flow_id, item_id,node_id, is_active, life)
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

// RetrieveAN gets the specified node from the database.
func RetrieveAN(ctx context.Context, db *sqlx.DB, nodeID, itemID, flowID string) (ActiveNode, error) {
	ctx, span := trace.StartSpan(ctx, "internal.flow.activeNode.Retrive")
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
	ctx, span := trace.StartSpan(ctx, "internal.flow.activeNode.ActiveNodes")
	defer span.End()

	activeNodes := []ActiveNode{}
	const q = `SELECT * FROM active_nodes where flow_id = any($1)`

	if err := db.SelectContext(ctx, &activeNodes, q, pq.Array(flowIDs)); err != nil {
		return activeNodes, errors.Wrap(err, "selecting active nodes")
	}

	return activeNodes, nil
}

func (an ActiveNode) entryNodeTrigger(ctx context.Context, db *sqlx.DB, n node.Node, entityID, itemID string, flowType int) error {
	if an.Life != 0 { //skips trigger if already has a life. TODO: It should be based on node condition as well
		return nil
	}
	return startNodeFlow(ctx, db, n, entityID, itemID, flowType)
}

func startNodeFlow(ctx context.Context, db *sqlx.DB, n node.Node, entityID, itemID string, flowType int) error {
	n.Variables = node.VariablesJSON(map[string]interface{}{entityID: itemID})
	n.Meta = node.Meta{
		EntityID: entityID,
		ItemID:   itemID,
		FlowType: flowType,
	}
	return prepareNextRun(ctx, db, n, map[string]interface{}{})
}

func startJobFlow(ctx context.Context, db *sqlx.DB, flowID, entityID, itemID string, flowType int) error {
	rootNode := node.Node{
		ID:        node.Root,
		FlowID:    flowID,
		Variables: node.VariablesJSON(map[string]interface{}{entityID: itemID}), //start with the item which triggered the flow
		Meta: node.Meta{
			EntityID: entityID,
			ItemID:   itemID,
			FlowType: flowType,
		},
	}
	return prepareNextRun(ctx, db, rootNode, map[string]interface{}{})
}

func prepareNextRun(ctx context.Context, db *sqlx.DB, n node.Node, parentResponseMap map[string]interface{}) error {
	logActiveNode(ctx, db, n)
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
	updatedParentResponseMap := updateVarJSON(n.VariablesMap(), parentResponseMap)
	for _, childNode := range childNodes {
		childNode.Meta = n.Meta //passing root node meta
		childNode.Variables = updatedParentResponseMap
		if childNode.Type == node.Stage { //stage nodes should not execute automatically. Always needs a manual intervention
			continue
		}
		// TODO call this in a job queue
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
	return prepareNextRun(ctx, db, n, ruleResult.Response)
}

func logActiveNode(ctx context.Context, db *sqlx.DB, n node.Node) error {
	log.Printf(">>>>>>>>>>>>>>>>        The Item Has Entered The Node Flow Node ID: %v -- Entity ID: %v -- Item ID: %v -- Flow ID: %v", n.ID, n.Meta.EntityID, n.Meta.ItemID, n.FlowID)
	if n.IsRootNode() {
		logFlowEvent(ctx, db, n)
		return nil
	}

	//Update the flow with the current active node.
	if n.Meta.FlowType != FlowTypePipeline {
		err := UpdateAFNode(ctx, db, n.ID, n.Meta.ItemID, n.FlowID)
		if err != nil {
			return err
		}
	}

	//Entry Active Node
	an := ActiveNode{
		AccountID: n.AccountID,
		FlowID:    n.FlowID,
		NodeID:    n.ID,
		ItemID:    n.Meta.ItemID,
		IsActive:  true,
		Life:      1,
	}
	_, err := CreateAN(ctx, db, an)
	return err
}

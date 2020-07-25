package flow

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/rule/engine"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"go.opencensus.io/trace"
)

// CreateAF inserts a new item into the active_flows table.
func CreateAF(ctx context.Context, db *sqlx.DB, af ActiveFlow) (ActiveFlow, error) {
	ctx, span := trace.StartSpan(ctx, "internal.rule.activeFlow.Create")
	defer span.End()

	const q = `INSERT INTO active_flows
		(flow_id, item_id, is_active, life)
		VALUES ($1, $2, $3, $4)`

	_, err := db.ExecContext(
		ctx, q,
		af.FlowID, af.ItemID, af.IsActive, af.Life,
	)
	if err != nil {
		return ActiveFlow{}, errors.Wrap(err, "inserting active_flow")
	}

	return af, nil
}

// UpdateAF modifies the active flow
func UpdateAF(ctx context.Context, db *sqlx.DB, af ActiveFlow) error {
	ctx, span := trace.StartSpan(ctx, "internal.rule.activeFlow.Update")
	defer span.End()

	const q = `UPDATE active_flows SET
		"is_active" = $3,
		"life" = $4
		WHERE item_id = $1 AND flow_id = $2`
	_, err := db.ExecContext(ctx, q, af.ItemID,
		af.FlowID, af.IsActive, af.Life,
	)
	if err != nil {
		return errors.Wrap(err, "updating active flow")
	}

	return nil
}

// activeFlows get the active flows entries for the dirty flow ids if exists
func activeFlows(ctx context.Context, lazyFlowsIDS []string, db *sqlx.DB) ([]ActiveFlow, error) {
	ctx, span := trace.StartSpan(ctx, "internal.rule.activeFlow.activeFlows")
	defer span.End()

	activeFlows := []ActiveFlow{}
	const q = `SELECT * FROM active_flows where flow_id = any($1)`

	if err := db.SelectContext(ctx, &activeFlows, q, pq.Array(lazyFlowsIDS)); err != nil {
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

func (af ActiveFlow) entryTrigger(ctx context.Context, db *sqlx.DB, flowID, entityID, itemID string, allowEntryTrigger bool) error {
	var err error
	if af.IsActive { //skips trigger if already active or of exit condition
		return nil
	}
	af.FlowID = flowID
	af.ItemID = itemID
	af.IsActive = true
	af.Life = af.Life + 1
	if af.Life == 1 {
		_, err = CreateAF(ctx, db, af)
	} else {
		err = UpdateAF(ctx, db, af)
	}
	if err == nil && allowEntryTrigger {
		err = startJobFlow(ctx, db, flowID, entityID, itemID)
	}

	return err
}

func (af ActiveFlow) exitTrigger(ctx context.Context, db *sqlx.DB, flowID, entityID, itemID string, allowExitTrigger bool) error {
	if af.Life == 0 || !af.IsActive { //skips trigger if new  or inactive.
		return nil
	}
	af.IsActive = false
	err := UpdateAF(ctx, db, af)

	if err == nil && allowExitTrigger {
		err = startJobFlow(ctx, db, flowID, entityID, itemID)
	}
	return err
}

func startJobFlow(ctx context.Context, db *sqlx.DB, flowID, entityID, itemID string) error {
	rootNode := node.Node{
		ID:        "root",
		FlowID:    flowID,
		Variables: node.VariablesJSON(map[string]interface{}{entityID: itemID}), //start with the item which triggered the flow
		Meta: node.Meta{
			EntityID: entityID,
			ItemID:   itemID,
		},
	}
	logFlowEvent(ctx, db, rootNode)
	return prepareNextRun(ctx, db, rootNode, map[string]interface{}{})
}

func prepareNextRun(ctx context.Context, db *sqlx.DB, n node.Node, parentResponseMap map[string]interface{}) error {
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
		childNode.Variables = updatedParentResponseMap
		// TODO call this in a job queue
		runJob(ctx, db, childNode)
	}
	return nil
}

func runJob(ctx context.Context, db *sqlx.DB, n node.Node) {
	engineRes, err := engine.RunRuleEngine(ctx, db, n)
	if err != nil {
		//TODO push this to DL queue
		return
	}
	logJobEvent(ctx, db, n)
	prepareNextRun(ctx, db, n, engineRes)
}

func updateVarJSON(existingVars map[string]interface{}, engineRes map[string]interface{}) string {
	for key, val := range existingVars {
		engineRes[key] = val
	}
	return node.VariablesJSON(engineRes)
}

func logFlowEvent(ctx context.Context, db *sqlx.DB, n node.Node) {
	log.Printf(">>>>>>>>>>>>>>>>        The Item Has Entered The Segment Flow %v", n.FlowID)
}

func logJobEvent(ctx context.Context, db *sqlx.DB, n node.Node) {
	log.Printf(">>>>>>>>>>>>>>>>        The Item Has Entered The Node Flow %v", n.ID)
}

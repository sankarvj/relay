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
	ctx, span := trace.StartSpan(ctx, "internal.rule.flow.activeFlows")
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
		createNodeJobs(ctx, db, flowID, entityID, itemID)
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
		createNodeJobs(ctx, db, flowID, entityID, itemID)
	}
	return err
}

func createNodeJobs(ctx context.Context, db *sqlx.DB, flowID, entityID, itemID string) {
	nodes, _ := node.List(ctx, flowID, db)
	log.Println("nodes ", nodes)
	branchNodeMap := node.BranceNodeMap(nodes)
	rootNode, err := node.RootNode(branchNodeMap)
	log.Printf("The rootNode %v", rootNode)
	log.Println("The rootNode err", err)

	rootNode.Variables = rootNode.VariablesJSONS(map[string]string{entityID: itemID})
	log.Printf("rootNode.Variables1 >>>>>>>>>>>>>>>>>>>>>>>>>>>>>> %s", rootNode.Variables)
	runJob(ctx, db, rootNode)
}

func runJob(ctx context.Context, db *sqlx.DB, n node.Node) {
	log.Printf("Running Job >>>>>>>>>>>>>>>>>>>>>>>>>>>>>> %s", n.ID)
	engineRes, err := engine.RunRuleEngine(ctx, db, n)
	if err != nil {
		//TODO push this is DL queue
		log.Println("Error running a job....", err)
	}

	nodes, _ := node.List(ctx, n.FlowID, db)
	childNodes, err := node.ChildNodes(n.ID, node.BranceNodeMap(nodes))

	//if multiple child nodes exists then who will take the job?
	//if the parentNode is a decision node than the result of engine.RunRuleEngine should say result:true/result:false
	//if the parentNode is a hook node than the the result of engine.RunRuleEngine should pass the API response inside the variables
	//if the parentNode is a push/modify/email node than the result of engine.RunRuleEngine should say result:true/result:false

	for _, childNode := range childNodes {
		childNode.Variables = childNode.VariablesJSON(engineRes)
		log.Printf("childNode.Variables >>>>>>>>>>>>>>>>>>>>>>>>>>>>>> %s", childNode.Variables)
		runJob(ctx, db, childNode)
	}

}

package flow

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
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

//ActiveFlows get the active flows entry for the lazy flows
func activeFlows(ctx context.Context, lazyFlowsIDS []string, db *sqlx.DB) ([]ActiveFlow, error) {
	ctx, span := trace.StartSpan(ctx, "internal.rule.flow.ActivateFlows")
	defer span.End()

	activeFlows := []ActiveFlow{}
	const q = `SELECT * FROM active_flows where flow_id = any($1)`

	if err := db.SelectContext(ctx, &activeFlows, q, pq.Array(lazyFlowsIDS)); err != nil {
		return activeFlows, errors.Wrap(err, "selecting active flows")
	}

	return activeFlows, nil
}

func activeFlowMap(activeFlows []ActiveFlow) map[string]ActiveFlow {
	activeFlowMap := map[string]ActiveFlow{}
	for _, aflow := range activeFlows {
		activeFlowMap[aflow.FlowID] = aflow
	}
	return activeFlowMap
}

func (af ActiveFlow) entryTrigger(ctx context.Context, db *sqlx.DB, itemID string, lf Flow) error {
	var err error
	if af.IsActive { //skips trigger if already active or of exit condition
		return nil
	}
	af.FlowID = lf.ID
	af.ItemID = itemID
	af.IsActive = true
	af.Life = af.Life + 1
	if af.Life == 1 {
		_, err = CreateAF(ctx, db, af)
	} else {
		err = UpdateAF(ctx, db, af)
	}

	if err == nil && lf.allowFlowEntryTrigger() {
		createNodeJobs(ctx, db, lf)
	}

	return err
}

func (af ActiveFlow) exitTrigger(ctx context.Context, db *sqlx.DB, lf Flow) error {
	if af.Life == 0 || !af.IsActive { //skips trigger if new  or inactive.
		return nil
	}
	af.IsActive = false
	err := UpdateAF(ctx, db, af)

	if err == nil && lf.allowFlowExitTrigger() {
		createNodeJobs(ctx, db, lf)
	}
	return err
}

func createNodeJobs(ctx context.Context, db *sqlx.DB, lf Flow) {
	nodes, _ := node.List(ctx, lf.ID, db)
	log.Println("nodes ", nodes)
	branchNodeMap := node.BranceNodeMap(nodes)
	rootNode, err := node.RootNode(branchNodeMap)
	log.Printf("The rootNode1 %v", rootNode)
	log.Println("The rootNode err", err)

	childNodes, err := node.ChildNodes(rootNode.ID, branchNodeMap)
	log.Printf("The childNodes %v", childNodes)
	log.Println("The childNodes err", err)

	//-------------------------------
	// for i, n := range nodes {
	// 	log.Printf("node %d -- %v", i, n)
	// 	engine.RunRuleEngine(tests.Context(), db, n)
	// }
}

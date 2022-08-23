package flow

import (
	"context"
	"database/sql"
	"encoding/json"
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
func List(ctx context.Context, entityIDs []string, fm int, ft int, db *sqlx.DB) ([]Flow, error) {
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

	types := []int{ft}
	if ft == FlowTypeAll {
		types = []int{FlowTypeUnknown, FlowTypeEntersSegment, FlowTypeLeavesSegment, FlowTypeEventCreate, FlowTypeEventUpdate}
	}

	q, args, err := sqlx.In(`SELECT * FROM flows where entity_id IN (?) AND mode IN (?) AND type IN (?) LIMIT 100;`, entityIDs, modes, types)
	if err != nil {
		return nil, errors.Wrap(err, "selecting flows")
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

	tokenBytes, err := json.Marshal(nf.Tokens)
	if err != nil {
		return Flow{}, errors.Wrap(err, "encode tokens to bytes")
	}
	tokenB := string(tokenBytes)

	f := Flow{
		ID:          nf.ID,
		AccountID:   nf.AccountID,
		EntityID:    nf.EntityID,
		Name:        nf.Name,
		Description: nf.Description,
		Expression:  nf.Expression,
		Tokenb:      &tokenB,
		Mode:        nf.Mode,
		Type:        nf.Type,
		Condition:   nf.Condition,
		Status:      FlowStatusActive,
		CreatedAt:   now.UTC(),
		UpdatedAt:   now.UTC().Unix(),
	}

	const q = `INSERT INTO flows
		(flow_id, account_id, entity_id, name, description, expression, tokenb, type, mode, condition, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err = db.ExecContext(
		ctx, q,
		f.ID, f.AccountID, f.EntityID, f.Name, f.Description, f.Expression, f.Tokenb, f.Type, f.Mode, f.Condition, f.Status,
		f.CreatedAt, f.UpdatedAt,
	)
	if err != nil {
		return Flow{}, errors.Wrap(err, "inserting flow")
	}

	return f, nil
}

func Update(ctx context.Context, db *sqlx.DB, uf NewFlow, now time.Time) (Flow, error) {
	ctx, span := trace.StartSpan(ctx, "internal.rule.flow.Update")
	defer span.End()

	const q = `UPDATE flows SET
		"entity_id" = $2,
		"name" = $3,
		"description" = $4,
		"type" = $5,
		"mode" = $6,
		"expression" = $7,
		"updated_at" = $8 
		 WHERE flow_id = $1`

	_, err := db.ExecContext(ctx, q, uf.ID,
		uf.EntityID, uf.Name, uf.Description, uf.Type, uf.Mode, uf.Expression, now.Unix(),
	)
	if err != nil {
		return Flow{}, errors.Wrap(err, "updating flow")
	}

	return Retrieve(ctx, uf.ID, db)
}

func UpdateStatus(ctx context.Context, db *sqlx.DB, accountID, flowID string, status int, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.rule.flow.UpdateStatus")
	defer span.End()

	const q = `UPDATE flows SET
		"status" = $1 
		 WHERE account_id = $2 AND flow_id = $3`

	_, err := db.ExecContext(ctx, q, status,
		accountID, flowID,
	)
	if err != nil {
		return errors.Wrap(err, "updating flow status failed")
	}

	return nil
}

//Call this only for segments
func (f *Flow) UpdateToken(ctx context.Context, db *sqlx.DB, tokens map[string]interface{}) error {
	ctx, span := trace.StartSpan(ctx, "internal.item.UpdateMeta")
	defer span.End()

	existingTokens := f.Tokens()
	for key, value := range tokens {
		existingTokens[key] = value
	}
	input, err := json.Marshal(existingTokens)
	if err != nil {
		return errors.Wrap(err, "encode meta to input")
	}
	tokenb := string(input)
	f.Tokenb = &tokenb

	const q = `UPDATE flows SET
		"toeknb" = $3 
		WHERE account_id = $1 AND entity_id = $2`
	_, err = db.ExecContext(ctx, q, f.AccountID, f.ID,
		f.Tokenb,
	)
	return err
}

func (f *Flow) Tokens() map[string]interface{} {
	display := make(map[string]interface{}, 0)
	if f.Tokenb == nil || *f.Tokenb == "" {
		return display
	}
	if err := json.Unmarshal([]byte(*f.Tokenb), &display); err != nil {
		log.Printf("***> unexpected error occurred when unmarshalling token for flow: %v error: %v\n", f.ID, err)
	}
	return display
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

func SearchByKey(ctx context.Context, accountID, entityID, term string, db *sqlx.DB) ([]Flow, error) {
	ctx, span := trace.StartSpan(ctx, "internal.flow.SearchByKey")
	defer span.End()

	flows := []Flow{}

	if term != "" {
		var q = `SELECT * FROM flows where account_id = $1 AND entity_id = $2 AND mode = $3 AND name LIKE '%` + term + `%'`

		if err := db.SelectContext(ctx, &flows, q, accountID, entityID, FlowModePipeLine); err != nil {
			return nil, errors.Wrap(err, "searching flows with search key")
		}
	}

	if len(flows) == 0 {
		var q = `SELECT * FROM flows where account_id = $1 AND entity_id = $2 AND mode = $3 LIMIT 100`

		if err := db.SelectContext(ctx, &flows, q, accountID, entityID, FlowModePipeLine); err != nil {
			return nil, errors.Wrap(err, "searching flows without search key")
		}
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
	_, span := trace.StartSpan(ctx, "internal.rule.flow.LazyFlows")
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

		if f.Status == FlowStatusInActive {
			log.Println("internal.rule.flow Trigger : Flow status in-active. Not going forward ")
			continue
		}

		af := activeFlowMap[f.ID]
		n := node.RootNode(f.AccountID, f.ID, f.EntityID, itemID, f.Expression).UpdateMeta(f.EntityID, itemID, f.Type).UpdateVariables(f.EntityID, itemID)
		if eng.RunExpEvaluator(ctx, db, rp, n.AccountID, n.Expression, n.VariablesMap()) { //entry
			if af.stopEntryTriggerFlow(f.Type) { //skip trigger if already active or of exit condition
				err = ErrFlowActive
			} else {
				err = af.entryFlowTrigger(ctx, db, rp, n, eng)
			}
		} else if f.Type == FlowTypeLeavesSegment {
			if af.stopExitTriggerFlow(f.Type) { //skip trigger if new or inactive or not allowed( i.e ftype != segment).
				err = ErrFlowInActive
			} else {
				err = af.exitFlowTrigger(ctx, db, rp, n, eng)
			}
		}
		//concat errors in the loop. nil if no error exists
		err = errors.Wrapf(err, "error in entry/exit trigger for flowID %q", f.ID)
		triggerErrors = append(triggerErrors, err)
	}

	if len(triggerErrors) > 0 {
		for _, err := range triggerErrors {
			if err != nil {
				log.Println("***> unexpected error occurred on trigger. error: ", err)
			}
		}
	}

	return triggerErrors
}

//DirectTrigger is when you want to execute the item on a particular node stage.
func DirectTrigger(ctx context.Context, db *sqlx.DB, rp *redis.Pool, accountID, flowID, nodeID, entityID, itemID string, eng engine.Engine) error {
	log.Printf("internal.rule.flow Direct Trigger : flowID:%s, nodeID:%s, entityID:%s, itemID:%s\n", flowID, nodeID, entityID, itemID)
	//retrival of primary components item,flow,node
	f, err := Retrieve(ctx, flowID, db)
	if err != nil {
		return err
	}

	if f.Status == FlowStatusInActive {
		log.Println("internal.rule.flow Direct Trigger : Flow status in-active. Not going forward ")
		return nil
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

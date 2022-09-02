package node

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

//Root node
const (
	Root = "00000000-0000-0000-0000-000000000000"
)

//Root node
const (
	NoActor = "00000000-0000-0000-0000-000000000000"
)

//GlobalEntity is the generic entity-id for certain expressions. See worker for its usecases
const (
	SelfEntity         = "self"
	SegmentEntity      = "segment"
	MeEntity           = "me"
	EmailEntityData    = "email"
	GlobalEntity       = "xyz"
	GlobalEntityData   = "data"
	GlobalEntityResult = "result"
	NoEntity           = "00000000-0000-0000-0000-000000000000"
)

var (
	// ErrNodeNotFound is used when a specific node is requested but does not exist.
	ErrNodeNotFound = errors.New("Node not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrInvalidNodeType occurs when the direct trigger executed for any node but stage-node
	ErrInvalidNodeType = errors.New("This operation is cannot be performed for this node type")
)

// List retrieves a list of existing nodes for the flow.
func List(ctx context.Context, flowID string, db *sqlx.DB) ([]Node, error) {
	ctx, span := trace.StartSpan(ctx, "internal.node.List")
	defer span.End()

	nodes := []Node{}
	const q = `SELECT * FROM nodes where flow_id = $1`

	if err := db.SelectContext(ctx, &nodes, q, flowID); err != nil {
		return nil, errors.Wrap(err, "selecting nodes")
	}

	return nodes, nil
}

// Stages retrieves a list of existing stages for the flow.
func Stages(ctx context.Context, accountID string, flowIDs []string, term string, db *sqlx.DB) ([]Node, error) {
	ctx, span := trace.StartSpan(ctx, "internal.node.Stages")
	defer span.End()

	nodes := []Node{}
	if term != "" {
		var q = `SELECT * FROM nodes where account_id = $1 AND flow_id = any($2) AND type = $3 AND name LIKE '%` + term + `'`

		if err := db.SelectContext(ctx, &nodes, q, accountID, pq.Array(flowIDs), Stage); err != nil {
			return nil, errors.Wrap(err, "selecting stages")
		}
	}

	if len(nodes) == 0 {
		const q = `SELECT * FROM nodes where account_id = $1 AND flow_id = any($2) AND type = $3 LIMIT 100`

		if err := db.SelectContext(ctx, &nodes, q, accountID, pq.Array(flowIDs), Stage); err != nil {
			return nil, errors.Wrap(err, "selecting stages")
		}
	}

	return nodes, nil
}

func BulkRetrieve(ctx context.Context, ids []interface{}, db *sqlx.DB) ([]Node, error) {
	ctx, span := trace.StartSpan(ctx, "internal.node.BulkRetrieveStages")
	defer span.End()

	nodes := []Node{}
	const q = `SELECT * FROM nodes where node_id = any($1) AND type = $2`

	if err := db.SelectContext(ctx, &nodes, q, pq.Array(ids), Stage); err != nil {
		return nodes, errors.Wrap(err, "selecting bulk stages for flow id")
	}

	return nodes, nil
}

//NodeActorsList is list with entity details joined
func NodeActorsList(ctx context.Context, flowID string, db *sqlx.DB) ([]NodeActor, error) {
	ctx, span := trace.StartSpan(ctx, "internal.node.NodeActorsList")
	defer span.End()

	nodes := []NodeActor{}
	const q = `select e.name as entity_name,e.category,n.node_id,n.flow_id,n.parent_node_id,n.actor_id,n.stage_id,n.name,n.description,n.weight,n.type,n.expression,n.tokenb,n.actuals from nodes as n left join entities as e on n.actor_id = e.entity_id where n.flow_id = $1`

	if err := db.SelectContext(ctx, &nodes, q, flowID); err != nil {
		return nil, errors.Wrap(err, "selecting node actors")
	}

	return nodes, nil
}

// Create inserts a new node into the database.
func Create(ctx context.Context, db *sqlx.DB, nn NewNode, now time.Time) (Node, error) {
	ctx, span := trace.StartSpan(ctx, "internal.node.Create")
	defer span.End()

	actuals, err := MapToJSONB(nn.Actuals)
	if err != nil {
		return Node{}, errors.Wrap(err, "encode actuals to bytes")
	}

	tokens, err := MapToJSONB(nn.Tokens)
	if err != nil {
		return Node{}, errors.Wrap(err, "encode tokens to bytes")
	}

	n := Node{
		ID:           nn.ID,
		ParentNodeID: nn.ParentNodeID,
		AccountID:    nn.AccountID,
		FlowID:       nn.FlowID,
		ActorID:      nn.ActorID,
		StageID:      nn.StageID,
		Name:         nn.Name,
		Description:  nn.Description,
		Weight:       nn.Weight,
		Type:         nn.Type,
		Expression:   nn.Expression,
		Tokenb:       tokens,
		Actuals:      actuals,
		CreatedAt:    now.UTC(),
		UpdatedAt:    now.UTC().Unix(),
	}

	const q = `INSERT INTO nodes
		(node_id, parent_node_id, account_id, flow_id, actor_id, stage_id, name, description, weight, type, expression, tokenb, actuals, 
		created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	_, err = db.ExecContext(
		ctx, q,
		n.ID, n.ParentNodeID, n.AccountID, n.FlowID, n.ActorID, n.StageID, n.Name, n.Description, n.Weight, n.Type, n.Expression, n.Tokenb, n.Actuals,
		n.CreatedAt, n.UpdatedAt,
	)
	if err != nil {
		return Node{}, errors.Wrap(err, "inserting node")
	}

	return n, nil
}

// Update replaces just the name all other fields are not updatable currenlty.
func Update(ctx context.Context, db *sqlx.DB, accountID, flowID, nodeID, name, expression string, tokens map[string]interface{}, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.node.Update")
	defer span.End()
	updatedAt := now.Unix()

	tokenb, err := MapToJSONB(tokens)
	if err != nil {
		return errors.Wrap(err, "encode tokens to bytes")
	}

	const q = `UPDATE nodes SET
		"name" = $4,
		"expression" = $5,
		"tokenb" = $6,
		"updated_at" = $7
		WHERE account_id = $1 AND flow_id = $2 AND node_id = $3`
	_, err = db.ExecContext(ctx, q, accountID, flowID, nodeID,
		name, expression, tokenb, updatedAt,
	)
	if err != nil {
		return errors.Wrap(err, "updating node")
	}

	return nil
}

func Map(ctx context.Context, db *sqlx.DB, accountID, flowID string, nm map[string]string, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.node.Map")
	defer span.End()

	//TODO do batch update by directly querring the psql instead of ORM way
	for k, v := range nm {
		const q = `UPDATE nodes SET
		"parent_node_id" = $4 
		WHERE account_id = $1 AND flow_id = $2 AND node_id = $3`
		_, err := db.ExecContext(ctx, q, accountID, flowID, k, v)
		if err != nil {
			return errors.Wrap(err, "updating node")
		}
	}

	return nil
}

// Retrieve gets the specified node from the database.
func Retrieve(ctx context.Context, accountID, flowID, id string, db *sqlx.DB) (*Node, error) {
	ctx, span := trace.StartSpan(ctx, "internal.node.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidID
	}

	var n Node
	const q = `SELECT * FROM nodes WHERE account_id = $1 AND flow_id = $2 AND node_id = $3`
	if err := db.GetContext(ctx, &n, q, accountID, flowID, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNodeNotFound
		}

		return nil, errors.Wrapf(err, "selecting node %q", id)
	}

	return &n, nil
}

//ChildNodes finds the child nodes
func ChildNodes(nodeID string, branchNodeMap map[string][]Node) []Node {
	for parentNodeID, nodes := range branchNodeMap {
		if parentNodeID == nodeID {
			return nodes
		}
	}
	return []Node{}
}

//MapToJSONB converts map to JSONB
func MapToJSONB(data interface{}) (string, error) {
	fieldsBytes, err := json.Marshal(data)
	if err != nil {
		return "", errors.Wrap(err, "encode data to bytes")
	}
	return string(fieldsBytes), err
}

//BranceNodeMap branches the node with the node id
func BranceNodeMap(nodes []Node) map[string][]Node {
	nodesBranchMap := map[string][]Node{}
	for _, node := range nodes {
		if existingNodes, ok := nodesBranchMap[node.ParentNodeID]; ok {
			nodesBranchMap[node.ParentNodeID] = append(existingNodes, node)
		} else {
			nodesBranchMap[node.ParentNodeID] = []Node{node}
		}
	}
	return nodesBranchMap
}

//VariablesJSON turns map to json
func VariablesJSON(varsMap map[string]interface{}) string {
	jsonStr, err := MapToJSONB(varsMap)
	if err != nil {
		log.Printf("***> unexpected error while marshalling node variables %v ", err)
		panic(err)
	}
	return jsonStr
}

//UpdateNodeVars updates the existing variables of the node with the new response map
func UpdateNodeVars(existingVars map[string]interface{}, newVars map[string]interface{}) map[string]interface{} {
	for key, exitingVal := range existingVars {
		if _, ok := newVars[key]; !ok { //if existing key present in newVars then keep the newVars value.
			newVars[key] = exitingVal
		} else if key == GlobalEntity || key == GlobalEntityData { // for global entity, the vars resides one level deeper
			// for the global entity, dive in to the innerMap (inside xyz).
			// we should update the content inside global entities and not just replace it with new values.
			exitingGlobalMap := existingVars[key].(map[string]interface{})
			newGlobalMap := newVars[key].(map[string]interface{})
			newVars[key] = UpdateNodeVars(exitingGlobalMap, newGlobalMap)
		}
	}
	return newVars
}

//IsRootNode decides whether the node is root or not
func (n Node) IsRootNode() bool {
	return n.ID == Root
}

//IsStageNode decides whether the node is stage type or not
func (n Node) IsStageNode() bool {
	return n.Type == Stage
}

//RootNode creates new root node from the flow
func RootNode(accountID, flowID, entityID, itemID, expression string) *Node {
	n := &Node{
		AccountID:  accountID,
		FlowID:     flowID,
		ID:         Root,
		Type:       Unknown,
		Actuals:    "",
		Expression: expression,
	}
	return n
}

//UpdateVariables initialize the variables of the node
//It puts the item which triggerred the flow into variables
func (n *Node) UpdateVariables(entityID, itemID string) *Node {
	n.Variables = VariablesJSON(UpdateNodeVars(n.VariablesMap(), map[string]interface{}{entityID: itemID}))
	return n
}

//UpdateVariables initialize the variables of the node
//It puts the item which triggerred the flow into variables
func (n *Node) UpdateNodeStage(entityID, itemID string) *Node {
	n.Variables = VariablesJSON(UpdateNodeVars(n.VariablesMap(), map[string]interface{}{entityID: itemID, "dd": n.StageID}))
	return n
}

//UpdateMeta updates the meta values of the node.
//Meta values includes the entity and its item which triggered this flow.
func (n *Node) UpdateMeta(entityID, itemID string, flowType int) *Node {
	n.Meta = Meta{
		EntityID: entityID,
		ItemID:   itemID,
		FlowType: flowType,
	}
	return n
}

func (n NodeActor) ActualsMap() map[string]string {
	var actualsMap map[string]string
	if err := json.Unmarshal([]byte(n.Actuals), &actualsMap); err != nil {
		//TODO handle this error properly
		errMsg := errors.Wrapf(err, "unexpected error occurred when unmarshalling node actuals attributes to actuals type %q", n.ID)
		log.Printf("***> unexpected error occurred when unmarshalling actualsMap for flow: %v error: %v\n", n.ID, errMsg)
		return actualsMap

	}
	return actualsMap
}

func (n NodeActor) Tokens() map[string]interface{} {
	display := make(map[string]interface{}, 0)
	if n.Tokenb == "" {
		return display
	}
	if err := json.Unmarshal([]byte(n.Tokenb), &display); err != nil {
		log.Printf("***> unexpected error occurred when unmarshalling token for flow: %v error: %v\n", n.ID, err)
	}
	return display
}

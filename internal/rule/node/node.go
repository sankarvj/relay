package node

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
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
func Stages(ctx context.Context, flowID string, db *sqlx.DB) ([]Node, error) {
	ctx, span := trace.StartSpan(ctx, "internal.node.Stages")
	defer span.End()

	nodes := []Node{}
	const q = `SELECT * FROM nodes where flow_id = $1 AND type = $2`

	if err := db.SelectContext(ctx, &nodes, q, flowID, Stage); err != nil {
		return nil, errors.Wrap(err, "selecting stages")
	}

	return nodes, nil
}

//NodeActorsList is list with entity details joined
func NodeActorsList(ctx context.Context, flowID string, db *sqlx.DB) ([]NodeActor, error) {
	ctx, span := trace.StartSpan(ctx, "internal.node.NodeActorsList")
	defer span.End()

	nodes := []NodeActor{}
	const q = `select e.name,e.category,n.node_id,n.parent_node_id,n.actor_id,n.type,n.expression,n.actuals from nodes as n left join entities as e on n.actor_id = e.entity_id where n.flow_id = $1`

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

	n := Node{
		ID:           nn.ID,
		ParentNodeID: nn.ParentNodeID,
		AccountID:    nn.AccountID,
		FlowID:       nn.FlowID,
		ActorID:      nn.ActorID,
		Name:         nn.Name,
		Type:         nn.Type,
		Expression:   nn.Expression,
		Actuals:      actuals,
		CreatedAt:    now.UTC(),
		UpdatedAt:    now.UTC().Unix(),
	}

	const q = `INSERT INTO nodes
		(node_id, parent_node_id, account_id, flow_id, actor_id, name, type, expression, actuals, 
		created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = db.ExecContext(
		ctx, q,
		n.ID, n.ParentNodeID, n.AccountID, n.FlowID, n.ActorID, n.Name, n.Type, n.Expression, n.Actuals,
		n.CreatedAt, n.UpdatedAt,
	)
	if err != nil {
		return Node{}, errors.Wrap(err, "inserting node")
	}

	return n, nil
}

// Retrieve gets the specified node from the database.
func Retrieve(ctx context.Context, id string, db *sqlx.DB) (*Node, error) {
	ctx, span := trace.StartSpan(ctx, "internal.node.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidID
	}

	var n Node
	const q = `SELECT * FROM nodes WHERE node_id = $1`
	if err := db.GetContext(ctx, &n, q, id); err != nil {
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
			existingNodes = append(existingNodes, node)
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
		log.Printf("error while marshalling node variables %v ", err)
		panic(err)
	}
	return jsonStr
}

//UpdateNodeVars updates the existing variables of the node with the new response map
func UpdateNodeVars(existingVars map[string]interface{}, newVars map[string]interface{}) map[string]interface{} {
	for key, exitingVal := range existingVars {
		if _, ok := newVars[key]; !ok { //if existing key present in newVars then keep the newVars value.
			newVars[key] = exitingVal
		} else if key == GlobalEntity || key == GlobalEntityData {
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

//UpdateMeta updates the meta values of the node
func (n *Node) UpdateMeta(entityID, itemID string, flowType int) *Node {
	n.Variables = VariablesJSON(UpdateNodeVars(n.VariablesMap(), map[string]interface{}{entityID: itemID})) //start with the item which triggered the flow
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
		errMsg := errors.Wrapf(err, "error while unmarshalling node actuals attributes to actuals type %q", n.ID)
		log.Println(errMsg)
		return actualsMap

	}
	return actualsMap
}

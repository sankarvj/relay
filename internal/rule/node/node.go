package node

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

var (
	// ErrNotFound is used when a specific node is requested but does not exist.
	ErrNotFound = errors.New("Node not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")
)

// List retrieves a list of existing nodes for the flow.
func List(ctx context.Context, flowID string, db *sqlx.DB) ([]Node, error) {
	ctx, span := trace.StartSpan(ctx, "internal.node.List")
	defer span.End()

	nodes := []Node{}
	const q = `SELECT * FROM nodes where flow_id = $1`

	if err := db.SelectContext(ctx, &nodes, q, flowID); err != nil {
		return nil, errors.Wrap(err, "selecting items")
	}

	return nodes, nil
}

// Create inserts a new node into the database.
func Create(ctx context.Context, db *sqlx.DB, nn NewNode, now time.Time) (Node, error) {
	ctx, span := trace.StartSpan(ctx, "internal.node.Create")
	defer span.End()

	variables, err := MapToJSONB(nn.Variables)
	if err != nil {
		return Node{}, errors.Wrap(err, "encode variables to bytes")
	}

	actuals, err := MapToJSONB(nn.Actuals)
	if err != nil {
		return Node{}, errors.Wrap(err, "encode variables to bytes")
	}

	n := Node{
		ID:           uuid.New().String(),
		ParentNodeID: nil,
		AccountID:    nn.AccountID,
		FlowID:       nn.FlowID,
		ActorID:      nn.ActorID,
		Type:         nn.Type,
		Expression:   nn.Expression,
		IsNeg:        nn.IsNeg,
		Variables:    variables,
		Actuals:      actuals,
		CreatedAt:    now.UTC(),
		UpdatedAt:    now.UTC().Unix(),
	}

	const q = `INSERT INTO nodes
		(node_id, parent_node_id, account_id, flow_id, actor_id, type, expression, is_neg, variables, actuals, 
		created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err = db.ExecContext(
		ctx, q,
		n.ID, n.ParentNodeID, n.AccountID, n.FlowID, n.ActorID, n.Type, n.Expression, n.IsNeg, n.Variables, n.Actuals,
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
	const q = `SELECT * FROM node WHERE node_id = $1`
	if err := db.GetContext(ctx, &n, q, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, errors.Wrapf(err, "selecting node %q", id)
	}

	return &n, nil
}

//RootNode fetches the root node
func RootNode(branchNodeMap map[string][]Node) (Node, error) {
	if rootNodes, ok := branchNodeMap["root"]; ok {
		if len(rootNodes) == 1 {
			return rootNodes[0], nil
		}
		return Node{}, errors.New("more than one root node exists")
	}
	return Node{}, errors.New("root node does not exists")
}

//ChildNodes finds the child nodes
func ChildNodes(nodeID string, branchNodeMap map[string][]Node) ([]Node, error) {
	for parentNodeID, nodes := range branchNodeMap {
		if parentNodeID == nodeID {
			return nodes, nil
		}
	}
	return []Node{}, errors.New("child node does not exists")
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
		if node.ParentNodeID == nil {
			nodesBranchMap["root"] = []Node{node}
			continue
		}
		if existingNodes, ok := nodesBranchMap[*node.ParentNodeID]; ok {
			existingNodes = append(existingNodes, node)
		} else {
			nodesBranchMap[*node.ParentNodeID] = []Node{node}
		}
	}
	return nodesBranchMap
}

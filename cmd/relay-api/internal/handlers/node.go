package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"go.opencensus.io/trace"
)

// Node represents the units inside the flow or inside the stage
type Node struct {
	db            *sqlx.DB
	rPool         *redis.Pool
	authenticator *auth.Authenticator
}

func (n *Node) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Node.Create")
	defer span.End()

	var nn node.NewNode
	if err := web.Decode(r, &nn); err != nil {
		return errors.Wrap(err, "")
	}
	nn.ID = uuid.New().String()
	nn = makeNode(params["account_id"], params["flow_id"], nn)

	log.Println("nn --> ", nn)
	//TODO: do it in single transaction <|>
	no, err := node.Create(ctx, n.db, nn, time.Now())
	if err != nil {
		return errors.Wrapf(err, "Node Create: %+v", &no)
	}

	return web.Respond(ctx, w, createViewModelNode(no), http.StatusCreated)
}

func (n *Node) Update(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Node.Update")
	defer span.End()

	var vn node.NewNode
	if err := web.Decode(r, &vn); err != nil {
		return errors.Wrap(err, "")
	}

	//currently it supports only the name update
	err := node.Update(ctx, n.db, params["account_id"], params["flow_id"], params["node_id"], vn.Name, time.Now())
	if err != nil {
		return errors.Wrapf(err, "Node Name Update: %+v", &vn)
	}

	return web.Respond(ctx, w, vn, http.StatusOK)
}

func (n *Node) Retrieve(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Node.Retrieve")
	defer span.End()

	no, err := node.Retrieve(ctx, params["account_id"], params["flow_id"], params["node_id"], n.db)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, createViewModelNode(*no), http.StatusOK)
}

func createViewModelNode(n node.Node) node.ViewModelNode {
	return node.ViewModelNode{
		ID:             n.ID,
		FlowID:         n.FlowID,
		StageID:        n.StageID,
		Name:           nameOfType(n.Type),
		Description:    n.Description,
		Expression:     n.Expression,
		ParentNodeID:   n.ParentNodeID,
		ActorID:        n.ActorID,
		EntityName:     "", //should I populate this?
		EntityCategory: -1, //should I populate this?
		Type:           n.Type,
		Actuals:        n.ActualsMap(),
	}
}

func makeNode(accountID, flowID string, nn node.NewNode) node.NewNode {
	//nn.ID = uuid.New().String() TODO:currently the view is creating it. Need to check
	nn.FlowID = flowID
	nn.AccountID = accountID
	if nn.ParentNodeID == "-1" || nn.ParentNodeID == "" {
		nn.ParentNodeID = node.Root
	}
	if nn.ActorID == "-1" || nn.ActorID == "" {
		nn.ActorID = node.NoActor
	}
	nn.Expression = makeExpression(nn.Queries)
	return nn
}

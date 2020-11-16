package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"go.opencensus.io/trace"
)

// Flow represents the journey
type Flow struct {
	db            *sqlx.DB
	rPool         *redis.Pool
	authenticator *auth.Authenticator
}

// List returns all the existing flows associated with entity
func (f *Flow) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Flow.List")
	defer span.End()

	var flows []flow.Flow
	if params["entity_id"] == "0" { //fetch all flows for all entities of the product if entity is zero
		entities, err := entity.List(ctx, params["team_id"], []int{}, f.db)
		if err != nil {
			return err
		}
		flows, err = flow.List(ctx, entity.FetchIDs(entities), f.db)
		if err != nil {
			return err
		}
	} else {
		var err error
		flows, err = flow.List(ctx, []string{params["entity_id"]}, f.db)
		if err != nil {
			return err
		}
	}

	viewModelFlows := make([]flow.ViewModelFlow, len(flows))
	for i, flow := range flows {
		viewModelFlows[i] = createViewModelFlow(flow, nil)
	}

	return web.Respond(ctx, w, viewModelFlows, http.StatusOK)
}

// Retrieve returns the specified flow from the system.
func (f *Flow) Retrieve(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Flow.Retrieve")
	defer span.End()

	fl, err := flow.Retrieve(ctx, params["flow_id"], f.db)
	if err != nil {
		return err
	}

	nodes, err := node.NodeActorsList(ctx, fl.ID, f.db)
	if err != nil {
		return err
	}

	viewModelNodes := make([]node.ViewModelNode, len(nodes))
	for i, node := range nodes {
		viewModelNodes[i] = createViewModelNode(node)
	}

	return web.Respond(ctx, w, createViewModelFlow(fl, viewModelNodes), http.StatusOK)
}

func (f *Flow) RetrieveActivedItems(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Flow.RetrieveActivedItems")
	defer span.End()

	fl, err := flow.Retrieve(ctx, params["flow_id"], f.db)
	if err != nil {
		return err
	}

	e, err := entity.Retrieve(ctx, params["entity_id"], f.db)
	if err != nil {
		return err
	}

	fields, err := e.Fields()
	if err != nil {
		return err
	}

	//TODO add pagination
	aflows, err := flow.ActiveFlows(ctx, []string{fl.ID}, f.db)
	if err != nil {
		return err
	}

	nodes, err := node.NodeActorsList(ctx, fl.ID, f.db)
	if err != nil {
		return err
	}

	viewModelNodes := make([]node.ViewModelNode, len(nodes))
	for i, node := range nodes {
		viewModelNodes[i] = createViewModelNode(node)
	}

	items, err := item.BulkRetrieve(ctx, e.ID, itemIds(aflows), f.db)
	if err != nil {
		return err
	}

	viewModelItems := make([]item.ViewModelItem, len(items))
	for i, item := range items {
		viewModelItems[i] = createViewModelItem(item)
	}

	response := struct {
		Items      []item.ViewModelItem `json:"items"`
		Flow       flow.ViewModelFlow   `json:"flow"`
		EntityName string               `json:"entity_name"`
		Fields     []entity.Field       `json:"fields"`
		Nodes      []node.ViewModelNode `json:"nodes"`
	}{
		Items:      viewModelItems,
		Flow:       createViewModelFlow(fl, nil),
		EntityName: e.Name,
		Fields:     fields,
		Nodes:      viewModelNodes,
	}

	return web.Respond(ctx, w, response, http.StatusOK)
}

func (f *Flow) RetrieveActiveNodes(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Flow.RetrieveActiveNodes")
	defer span.End()

	activeNodes, err := flow.ActiveNodesForItem(ctx, params["account_id"], params["flow_id"], params["item_id"], f.db)
	if err != nil {
		return err
	}

	viewModelActiveNodes := make([]node.ViewModelActiveNode, len(activeNodes))
	for i, an := range activeNodes {
		viewModelActiveNodes[i] = createViewModelActiveNode(an)
	}

	return web.Respond(ctx, w, viewModelActiveNodes, http.StatusOK)
}

// Create inserts a new flow into the entity.
func (f *Flow) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Flow.Create")
	defer span.End()

	var nf flow.NewFlow
	if err := web.Decode(r, &nf); err != nil {
		return errors.Wrap(err, "")
	}
	nf.ID = uuid.New().String()
	nf.AccountID = params["account_id"]
	nf.EntityID = params["entity_id"]

	//TODO: do it in single transaction <|>
	flow, err := flow.Create(ctx, f.db, nf, time.Now())
	if err != nil {
		return errors.Wrapf(err, "Flow: %+v", &flow)
	}

	for _, nn := range nf.Nodes {
		//TODO: do it in single transaction >|<
		nn = makeNode(flow.AccountID, flow.ID, nn)
		n, err := node.Create(ctx, f.db, nn, time.Now())
		if err != nil {
			return errors.Wrapf(err, "Node: %+v", n)
		}
	}

	return web.Respond(ctx, w, flow, http.StatusCreated)
}

func createViewModelFlow(f flow.Flow, nodes []node.ViewModelNode) flow.ViewModelFlow {
	return flow.ViewModelFlow{
		ID:          f.ID,
		EntityID:    f.EntityID,
		Name:        f.Name,
		Description: f.Description,
		Expression:  f.Expression,
		Nodes:       nodes,
	}
}

func makeNode(accountID, flowID string, nn node.NewNode) node.NewNode {
	nn.FlowID = flowID
	nn.AccountID = accountID
	if nn.ParentNodeID == "-1" || nn.ParentNodeID == "" {
		nn.ParentNodeID = node.Root
	}
	if nn.ActorID == "-1" || nn.ActorID == "" {
		nn.ActorID = node.NoActor
	}
	for i, q := range nn.Queries {
		nn.Expression = fmt.Sprintf("%s %s %s", q.Key, q.Operator, q.Value)
		if i < len(nn.Queries)-1 {
			nn.Expression = fmt.Sprintf("%s %s", nn.Expression, "AND")
		}
	}
	return nn
}

func itemIds(actFlows []flow.ActiveFlow) []string {
	ids := make([]string, len(actFlows))
	for i, aflow := range actFlows {
		ids[i] = aflow.ItemID
	}
	return ids
}

func entityIds(nodes []node.Node) []string {
	ids := make([]string, len(nodes))
	for i, n := range nodes {
		ids[i] = n.ActorID
	}
	return ids
}

func createViewModelNode(n node.NodeActor) node.ViewModelNode {
	return node.ViewModelNode{
		ID:             n.ID,
		Name:           nameOfType(n.Type),
		ParentNodeID:   n.ParentNodeID,
		ActorID:        n.ActorID,
		EntityName:     n.EntityName.String,
		EntityCategory: int(n.EntityCategory.Int32),
		Type:           n.Type,
		Actuals:        n.ActualsMap(),
	}
}

func createViewModelActiveNode(n flow.ActiveNode) node.ViewModelActiveNode {
	return node.ViewModelActiveNode{
		ID:       n.NodeID,
		IsActive: n.IsActive,
		Life:     n.Life,
	}
}

func nameOfType(typeOfNode int) string {
	//TODO: Remove it here. Hanlde this in the UI
	return ""
}

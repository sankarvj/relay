package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"go.opencensus.io/trace"
)

// Flow represents the journey
type Flow struct {
	db            *sqlx.DB
	authenticator *auth.Authenticator
}

// List returns all the existing flows associated with entity
func (f *Flow) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Flow.List")
	defer span.End()

	fm := flowMode(r.URL.Query().Get("fm"))

	var flows []flow.Flow
	entityIDs := []string{params["entity_id"]}
	if params["entity_id"] == "0" { //fetch all flows for all entities of the product if entity is zero
		entities, err := entity.List(ctx, params["account_id"], params["team_id"], []int{}, f.db)
		if err != nil {
			return err
		}
		entityIDs = entity.FetchIDs(entities)
	}

	flows, err := flow.List(ctx, entityIDs, fm, flow.FlowTypeAll, f.db)
	if err != nil {
		return err
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

	nodes, err := node.NodeActorsList(ctx, params["account_id"], fl.ID, f.db)
	if err != nil {
		return err
	}

	viewModelNodes := make([]node.ViewModelNode, len(nodes))
	for i, node := range nodes {
		viewModelNodes[i] = createViewModelNodeActor(node)
	}

	return web.Respond(ctx, w, createViewModelFlow(fl, viewModelNodes), http.StatusOK)
}

//remove this method. useful only for verification of flow path
func (f *Flow) FlowTrails(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Flow.FlowTrails")
	defer span.End()

	next := util.ConvertStrToInt(r.URL.Query().Get("next"))

	fl, err := flow.Retrieve(ctx, params["flow_id"], f.db)
	if err != nil {
		return err
	}

	e, err := entity.Retrieve(ctx, params["account_id"], params["entity_id"], f.db)
	if err != nil {
		return err
	}

	fields, err := e.Fields()
	if err != nil {
		return err
	}

	//TODO add pagination
	aflows, err := flow.ActiveFlows(ctx, params["account_id"], []string{fl.ID}, next, f.db)
	if err != nil {
		return err
	}

	nodes, err := node.NodeActorsList(ctx, params["account_id"], fl.ID, f.db)
	if err != nil {
		return err
	}

	viewModelNodes := make([]node.ViewModelNode, len(nodes))
	for i, node := range nodes {
		viewModelNodes[i] = createViewModelNodeActor(node)
	}

	var items []item.Item
	if len(aflows) > 0 {
		items, err = item.BulkRetrieve(ctx, e.ID, itemIds(aflows), f.db)
		if err != nil {
			return err
		}
	}

	viewModelItems := make([]ViewModelItem, len(items))
	for i, item := range items {
		viewModelItems[i] = createViewModelItem(item)
	}

	next = next + 1
	response := struct {
		Items      []ViewModelItem      `json:"items"`
		Flow       flow.ViewModelFlow   `json:"flow"`
		EntityName string               `json:"entity_name"`
		Fields     []entity.Field       `json:"fields"`
		Nodes      []node.ViewModelNode `json:"nodes"`
		Next       string               `json:"next"`
	}{
		Items:      viewModelItems,
		Flow:       createViewModelFlow(fl, nil),
		EntityName: e.DisplayName,
		Fields:     fields,
		Nodes:      viewModelNodes,
		Next:       fmt.Sprint(next),
	}

	return web.Respond(ctx, w, response, http.StatusOK)
}

func (f *Flow) TrailNodes(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Flow.TrailNodes")
	defer span.End()

	next := util.ConvertStrToInt(r.URL.Query().Get("next"))

	activeNodes, err := flow.ActiveNodesForItem(ctx, params["account_id"], params["flow_id"], params["item_id"], next, f.db)
	if err != nil {
		return err
	}

	viewModelActiveNodes := make([]node.ViewModelActiveNode, len(activeNodes))
	for i, an := range activeNodes {
		viewModelActiveNodes[i] = createViewModelActiveNode(an)
	}

	next = next + 1
	response := struct {
		Nodes []node.ViewModelActiveNode `json:"nodes"`
		Next  string                     `json:"next"`
	}{
		Nodes: viewModelActiveNodes,
		Next:  fmt.Sprint(next),
	}

	return web.Respond(ctx, w, response, http.StatusOK)
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
	nf.Expression, nf.Tokens = makeExpression(nf.Queries)

	flow, err := flow.Create(ctx, f.db, nf, time.Now())
	if err != nil {
		return errors.Wrapf(err, "error creating flow: %+v", &flow)
	}

	return web.Respond(ctx, w, createViewModelFlow(flow, []node.ViewModelNode{}), http.StatusCreated)
}

func (f *Flow) Update(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Flow.Update")
	defer span.End()

	var uf flow.NewFlow
	if err := web.Decode(r, &uf); err != nil {
		return errors.Wrap(err, "")
	}
	uf.Expression, uf.Tokens = makeExpression(uf.Queries)
	uf.AccountID = params["account_id"]
	uf.EntityID = params["entity_id"]

	uuf, err := flow.Update(ctx, f.db, uf, time.Now())
	if err != nil {
		return errors.Wrapf(err, "Error updating flow")
	}

	nodes, err := node.NodeActorsList(ctx, uf.AccountID, uuf.ID, f.db)
	if err != nil {
		return err
	}

	viewModelNodes := make([]node.ViewModelNode, len(nodes))
	for i, node := range nodes {
		viewModelNodes[i] = createViewModelNodeActor(node)
	}

	return web.Respond(ctx, w, createViewModelFlow(uuf, viewModelNodes), http.StatusOK)
}

func (f *Flow) UpdateStatus(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Flow.UpdateStatus")
	defer span.End()

	var uf flow.NewFlow
	if err := web.Decode(r, &uf); err != nil {
		return errors.Wrap(err, "")
	}

	err := flow.UpdateStatus(ctx, f.db, params["account_id"], params["flow_id"], uf.Status, time.Now())
	if err != nil {
		return errors.Wrapf(err, "Error updating flow status")
	}

	return web.Respond(ctx, w, nil, http.StatusOK)
}

func createViewModelFlow(f flow.Flow, nodes []node.ViewModelNode) flow.ViewModelFlow {
	return flow.ViewModelFlow{
		ID:          f.ID,
		EntityID:    f.EntityID,
		Name:        f.Name,
		Description: f.Description,
		Expression:  f.Expression,
		Mode:        f.Mode,
		State:       f.State,
		Type:        f.Type,
		Status:      f.Status,
		Nodes:       nodes,
		Tokens:      f.Tokens(),
	}
}

func itemIds(actFlows []flow.ActiveFlow) []interface{} {
	ids := make([]interface{}, len(actFlows))
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

func createViewModelNodeActor(n node.NodeActor) node.ViewModelNode {
	return node.ViewModelNode{
		ID:             n.ID,
		FlowID:         n.FlowID,
		StageID:        n.StageID,
		Name:           n.Name,
		Description:    n.Description,
		Expression:     n.Expression,
		ParentNodeID:   n.ParentNodeID,
		ActorID:        n.ActorID,
		Weight:         n.Weight,
		EntityName:     n.EntityName.String,
		EntityCategory: int(n.EntityCategory.Int32),
		Type:           n.Type,
		Tokens:         n.Tokens(),
		Actuals:        n.ActualsMap(),
	}
}

func createViewModelActiveNode(n flow.ActiveNode) node.ViewModelActiveNode {
	return node.ViewModelActiveNode{
		ID:        n.NodeID,
		IsActive:  n.IsActive,
		Life:      n.Life,
		CreatedAt: n.CreatedAt,
	}
}

func nameOfType(typeOfNode int) string {
	//TODO: Remove it here. Hanlde this in the UI
	return ""
}

func makeExpression(queries []node.Query) (string, map[string]interface{}) {
	var expression string
	tokens := make(map[string]interface{}, 0)
	for i, q := range queries {
		if q.Key == "" {
			continue
		}
		if q.Value != q.Display {
			tokens[q.Value] = q.Display
		}
		expression = fmt.Sprintf("%s {{%s.%s}} %s {%s}", expression, q.EntityID, q.Key, q.Operator, q.Value)
		if i < len(queries)-1 {
			expression = fmt.Sprintf("%s %s", expression, "&&")
		}
	}
	return expression, tokens
}

func flowMode(fm string) int {
	i, err := strconv.Atoi(fm)
	if err != nil {
		log.Printf("***> unexpected error occurred cannot parse fm from the request %s", err)
		return flow.FlowModeWorkFlow
	}
	return i
}

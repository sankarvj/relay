package handlers

import (
	"context"

	"github.com/jmoiron/sqlx"
	rg "github.com/redislabs/redisgraph-go"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

type Piper struct {
	Viable bool                       `json:"viable"`
	Pipe   bool                       `json:"pipe"`
	Flows  []flow.ViewModelFlow       `json:"flows"`
	Nodes  []node.ViewModelNode       `json:"nodes"`
	Items  map[string][]ViewModelItem `json:"items"`
}

func setRenderer(ctx context.Context, ls string, e entity.Entity, db *sqlx.DB) string {
	if ls == entity.MetaRenderPipe {
		if !e.IsPipeLayout() { //update pipe in entity
			e.UpdateMeta(ctx, db, map[string]interface{}{entity.MetaRender: entity.MetaRenderPipe})
		}
	} else if ls == entity.MetaRenderList {
		if e.IsPipeLayout() { //update pipe in entity
			e.UpdateMeta(ctx, db, map[string]interface{}{entity.MetaRender: entity.MetaRenderList})
		}
	} else {
		if e.IsPipeLayout() {
			ls = entity.MetaRenderPipe
		} else {
			ls = entity.MetaRenderList
		}
	}
	return ls
}

func pipeKanban(ctx context.Context, e entity.Entity, p *Piper, db *sqlx.DB) error {
	//If true, pass the values needed for the view
	var viewModelFlows []flow.ViewModelFlow
	var viewModelNodes []node.ViewModelNode
	if e.FlowField() != nil {
		flows, err := flow.List(ctx, []string{e.ID}, flow.FlowModePipeLine, flow.FlowTypeAll, db)
		if err != nil {
			return err
		}

		viewModelFlows = make([]flow.ViewModelFlow, len(flows))
		for i, flow := range flows {
			viewModelFlows[i] = createViewModelFlow(flow, nil)
		}

		if len(flows) > 0 {
			nodes, err := node.NodeActorsList(ctx, flows[0].ID, db)
			if err != nil {
				return err
			}

			viewModelNodes = make([]node.ViewModelNode, len(nodes))
			for i, n := range nodes {
				if n.Type == node.Stage {
					viewModelNodes[i] = createViewModelNodeActor(n)
				}
			}

		}

	}

	p.Flows = viewModelFlows
	p.Nodes = viewModelNodes
	p.Items = make(map[string][]ViewModelItem, 0)

	return nil
}

func makeGraphField(f *entity.Field, value interface{}, expression string) graphdb.Field {
	if f.IsReference() {
		return graphdb.Field{
			Key:       f.Key,
			Value:     []interface{}{""},
			DataType:  graphdb.TypeReference,
			RefID:     f.RefID,
			IsReverse: false,
			Field: &graphdb.Field{
				Expression: graphdb.Operator(expression),
				Key:        "id",
				DataType:   graphdb.TypeString,
				Value:      value,
			},
		}
	} else if f.IsList() {
		return graphdb.Field{
			Key:      f.Key,
			Value:    []interface{}{value},
			DataType: graphdb.DType(f.DataType),
			Field: &graphdb.Field{
				Expression: graphdb.Operator(expression),
				Key:        "element",
				DataType:   graphdb.DType(f.Field.DataType),
			},
		}
	} else {
		return graphdb.Field{
			Expression: graphdb.Operator(expression),
			Key:        f.Key,
			DataType:   graphdb.DType(f.DataType),
			Value:      value,
		}
	}
}

func itemsResp(ctx context.Context, db *sqlx.DB, accountID string, result *rg.QueryResult) ([]item.Item, error) {
	items, err := item.BulkRetrieveItems(ctx, accountID, itemIDs(result), db)
	if err != nil {
		return []item.Item{}, err
	}

	return items, nil
}

func itemIDs(result *rg.QueryResult) []interface{} {
	itemIds := make([]interface{}, 0)
	for result.Next() { // Next returns true until the iterator is depleted.
		// Get the current Record.
		r := result.Record()

		// Entries in the Record can be accessed by index or key.
		record := util.ConvertInterfaceToMap(util.ConvertInterfaceToMap(r.GetByIndex(0))["Properties"])
		itemIds = append(itemIds, record["id"])
	}
	return itemIds
}

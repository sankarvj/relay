package handlers

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	rg "github.com/redislabs/redisgraph-go"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"gitlab.com/vjsideprojects/relay/internal/team"
)

type Piper struct {
	Viable         bool                       `json:"viable"`
	Group          bool                       `json:"group"`
	LS             string                     `json:"layout_style"`
	NodeKey        string                     `json:"node_key"`
	Flows          []flow.ViewModelFlow       `json:"flows"`
	Nodes          []node.ViewModelNode       `json:"nodes"`
	Items          map[string][]ViewModelItem `json:"items"`
	Tokens         map[string]string          `json:"tokens"`
	Exps           map[string]string          `json:"exps"`
	CountMap       map[string]map[string]int  `json:"count_map"`
	sourceEntityID string
	sourceItemID   string
}

func setRenderer(ctx context.Context, ls string, e entity.Entity, db *sqlx.DB) string {
	if ls == entity.MetaRenderPipe && !e.IsPipeLayout() {
		log.Println("***> ***> ***>Set renderer1")
		err := e.UpdateMeta(ctx, db, map[string]interface{}{entity.MetaRender: entity.MetaRenderPipe})
		if err != nil {
			log.Println("***> unexpected error occurred in internal.handlers.item. when setting pipe renderer", err)
		}
	} else if ls == entity.MetaRenderList && e.IsPipeLayout() {
		log.Println("***> ***> ***>Set renderer2")
		err := e.UpdateMeta(ctx, db, map[string]interface{}{entity.MetaRender: entity.MetaRenderList})
		if err != nil {
			log.Println("***> unexpected error occurred in internal.handlers.item. when setting list renderer", err)
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

func pipeKanban(ctx context.Context, accountID string, e entity.Entity, p *Piper, db *sqlx.DB) error {
	//If true, pass the values needed for the view
	var viewModelFlows []flow.ViewModelFlow
	var viewModelNodes []node.ViewModelNode
	if e.FlowField() != nil { //main stages. ex: deal stages
		flows, err := flow.List(ctx, []string{e.ID}, flow.FlowModePipeLine, flow.FlowTypeAll, db)
		if err != nil {
			return err
		}

		viewModelFlows = make([]flow.ViewModelFlow, len(flows))
		for i, flow := range flows {
			viewModelFlows[i] = createViewModelFlow(flow, nil)
		}

		if len(flows) > 0 {
			viewModelNodes, err = nodeStages(ctx, flows[0].ID, db)
			if err != nil {
				return err
			}
		}

	} else if p.sourceEntityID != "" && p.sourceItemID != "" { //child stages. ex: tasks created in the deal stages
		e, err := entity.Retrieve(ctx, accountID, p.sourceEntityID, db)
		if err != nil {
			return err
		}
		it, err := item.Retrieve(ctx, p.sourceEntityID, p.sourceItemID, db)
		if err != nil {
			return err
		}
		f := e.FlowField()
		if f != nil {
			flowIDs := it.Fields()[f.Key]
			if flowIDs != nil {
				flowID := flowIDs.([]interface{})[0]
				viewModelNodes, err = nodeStages(ctx, flowID.(string), db)
				if err != nil {
					return err
				}
			}
		}
	}

	log.Println("viewModelNodes ", viewModelNodes)

	p.Flows = viewModelFlows
	p.Nodes = viewModelNodes
	p.Items = make(map[string][]ViewModelItem, 0)

	return nil
}

func nodeStages(ctx context.Context, flowID string, db *sqlx.DB) ([]node.ViewModelNode, error) {
	nodes, err := node.NodeActorsList(ctx, flowID, db)
	if err != nil {
		return nil, err
	}

	viewModelNodes := make([]node.ViewModelNode, 0)
	for _, n := range nodes {
		if n.Type == node.Stage {
			viewModelNodes = append(viewModelNodes, createViewModelNodeActor(n))
		}
	}
	return viewModelNodes, nil
}

func makeGraphField(f *entity.Field, value interface{}, expression string, reverse bool) graphdb.Field {
	if f.IsReference() {
		dataType := graphdb.TypeString
		if strings.EqualFold(lexertoken.INSign, expression) || strings.EqualFold(lexertoken.NotINSign, expression) {
			dataType = graphdb.TypeWist
			switch v := value.(type) {
			case string:
				arr := strings.Split(strings.ReplaceAll(v, " ", ""), ",")
				value = arr
			}
		}
		return graphdb.Field{
			Key:       f.Key,
			Value:     []interface{}{""},
			DataType:  graphdb.TypeReference,
			RefID:     f.RefID,
			IsReverse: reverse,
			Field: &graphdb.Field{
				Expression: graphdb.Operator(expression),
				Key:        "id",
				DataType:   dataType,
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
	} else if f.IsDateOrTime() { // populates min and max range with the time value. if `-` exists. Assumption: All the datetime segmentation values has this format start_time_in_millis-end_time_in_millis
		var min string
		var max string
		dataType := graphdb.DType(f.DataType)
		switch v := value.(type) {
		case string:
			if value == "now" {
				dataType = graphdb.TypeDateTimeMillis
				value = util.GetMilliSecondsStr(time.Now().UTC())
			} else {
				parts := strings.Split(v, "-")
				if len(parts) == 2 { // date range
					dataType = graphdb.TypeDateRange
					min = parts[0]
					max = parts[1]
				}
			}

		case int, int64:
			dataType = graphdb.TypeDateTimeMillis
		}

		return graphdb.Field{
			Expression: graphdb.Operator(expression),
			Key:        f.Key,
			DataType:   dataType,
			Value:      value,
			Min:        min,
			Max:        max,
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
	itemIDs := itemIDs(result)
	items, err := item.BulkRetrieveItems(ctx, accountID, itemIDs, db)
	if err != nil {
		return []item.Item{}, err
	}

	return sort(items, itemIDs), nil
}

func itemIDs(result *rg.QueryResult) []interface{} {
	itemIds := make([]interface{}, 0)
	if result != nil {
		for result.Next() { // Next returns true until the iterator is depleted.
			// Get the current Record.
			r := result.Record()

			// Entries in the Record can be accessed by index or key.
			record := util.ConvertInterfaceToMap(util.ConvertInterfaceToMap(r.GetByIndex(0))["Properties"])
			//if record["id"] != "--" { //TODO: hacky-none-fix
			itemIds = append(itemIds, record["id"])
			//}

		}
	}

	return itemIds
}

func itemElements(result *rg.QueryResult) []interface{} {
	values := make([]interface{}, 0)
	for result.Next() { // Next returns true until the iterator is depleted.
		// Get the current Record.
		r := result.Record()

		// Entries in the Record can be accessed by index or key.
		record := util.ConvertInterfaceToMap(util.ConvertInterfaceToMap(r.GetByIndex(0))["Properties"])
		log.Printf("record %+v", record)
		values = append(values, record["element"])
	}
	return values
}

func sort(items []item.Item, itemIds []interface{}) []item.Item {
	itemMap := make(map[string]item.Item, len(items))
	for i := 0; i < len(items); i++ {
		itemMap[items[i].ID] = items[i]
	}
	sortedItems := make([]item.Item, 0)
	for _, id := range itemIds {
		sortedItems = append(sortedItems, itemMap[id.(string)])
	}
	return sortedItems
}

func selectedTeam(ctx context.Context, accountID, teamID string, db *sqlx.DB) ([]team.Team, string, error) {
	teams, err := team.List(ctx, accountID, db)
	if err != nil {
		return nil, "", err
	}

	var oldTeamID string
	var seletedTeamID string
	for _, t := range teams {
		if t.ID == teamID {
			oldTeamID = t.ID
		}
		if t.ID != t.AccountID {
			seletedTeamID = t.ID
		}
	}
	if oldTeamID != "" {
		seletedTeamID = oldTeamID
	}
	return teams, seletedTeamID, nil
}

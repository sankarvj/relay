package reference

import (
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

type Choicer struct {
	ID   string
	Name string
}

func nodeChoices(nodes []node.Node) []Choicer {
	choicers := make([]Choicer, len(nodes))
	for i, node := range nodes {
		choicers[i] = Choicer{
			ID:   node.ID,
			Name: node.Name,
		}
	}
	return choicers
}

func flowChoices(flows []flow.Flow) []Choicer {
	choicers := make([]Choicer, len(flows))
	for i, flow := range flows {
		choicers[i] = Choicer{
			ID:   flow.ID,
			Name: flow.Name,
		}
	}
	return choicers
}

func itemChoices(f entity.Field, items []item.Item) []Choicer {
	choicers := make([]Choicer, len(items))
	for i, item := range items {
		choicers[i] = Choicer{
			ID:   item.ID,
			Name: item.Fields()[f.DisplayGex()].(string),
		}
	}
	return choicers
}

func isIdNotExistAlready(ids []string, newid string) bool {
	for _, id := range ids {
		if id == newid {
			return false
		}
	}
	return true
}
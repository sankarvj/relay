package reference

import (
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

type Choicer struct {
	ID     string
	Name   string
	Value  interface{}
	Verb   interface{}
	Avatar interface{}
	Color  interface{}
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

func ItemChoices(f *entity.Field, items []item.Item, whoMap map[string]string) []Choicer {
	choicers := make([]Choicer, len(items))
	for i, item := range items {
		var displayName interface{}
		displayNameStr := ""
		if f != nil {
			displayName = item.Fields()[f.DisplayGex()] // finding the lookup from the main field
		}

		if displayName != nil {
			displayNameStr = displayName.(string)
		} else if item.Name != nil {
			displayNameStr = *item.Name
		}
		choicers[i] = Choicer{
			ID:     item.ID,
			Name:   displayNameStr,
			Color:  item.Fields()[whoMap[entity.WhoColor]],
			Value:  item.Fields()[whoMap[entity.WhoIdentifier]], // considering it as view value...  is it okay to have a specific logic with name emailgex?
			Verb:   item.Fields()[whoMap[entity.WhoVerb]],       // is it okay to have `uuid-00-verb`?
			Avatar: item.Fields()[whoMap[entity.WhoAvatar]],     // finding the lookup from the child itemn
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

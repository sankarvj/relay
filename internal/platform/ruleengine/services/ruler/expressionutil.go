package ruler

import (
	"strconv"
	"strings"
)

//Behaviour is the enum to hold the action behaviour
type Behaviour string

//Types of behaviours
const (
	Create  Behaviour = "create"
	Update  Behaviour = "update"
	Retrive Behaviour = "retrive"
)

//Action is the parsed action expression
type Action struct {
	EntityID  string
	ItemID    string
	SecItemID string
	Behaviour Behaviour
}

//FetchEntityID fetches the root id from the rule
func FetchEntityID(expression string) string {
	parts := strings.Split(expression, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

//FetchItemID fetches item type from the list
func FetchItemID(expression string) string {
	parts := strings.Split(expression, ".")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

//Evaluate evaluates the expression with the coresponding map
func Evaluate(expression string, response map[string]interface{}) interface{} {
	var realValue interface{}
	elements := strings.Split(expression, ".")
	lenOfElements := len(elements)
	for index, element := range elements {
		if index == (lenOfElements - 1) {
			realValue = response[element]
			break
		}
		if response[element] == nil {
			break
		}
		response = response[element].(map[string]interface{})
	}
	return realValue
}

func extract(value string) interface{} {
	if v, err := strconv.Atoi(value); err == nil {
		return v
	}
	return value
}

package ruler

import (
	"strconv"
	"strings"
)

//FetchRootKey fetches the root id from the rule
func FetchRootKey(value string) string {
	parts := strings.Split(value, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

//FetchItemType fetches item type from the list
func FetchItemType(ruleVal string) string {
	parts := strings.Split(ruleVal, ".")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

func evaluate(expression string, response map[string]interface{}) interface{} {
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

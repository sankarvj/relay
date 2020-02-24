package ruler

import (
	"log"
	"reflect"
	"strconv"
	"strings"

	version "github.com/mcuadros/go-version"
)

func compare(left, right Operand) bool {
	log.Println("Actuals....")
	log.Println("left ", left)
	log.Println("right ", right)
	log.Println("Types....")
	log.Println("left ", reflect.TypeOf(left))
	log.Println("right ", reflect.TypeOf(right))
	if left == nil && right == nil {
		return false
	}

	//if the type is sting but version
	if version.ValidSimpleVersionFormat(left.(string)) && version.ValidSimpleVersionFormat(right.(string)) {
		r := version.CompareSimple(left.(string), right.(string))
		if r == 0 {
			return true
		}
		return false
	}

	return left == right
}

func evaluate(value string, response map[string]interface{}) interface{} {
	var realValue interface{}
	parts := strings.Split(value, ".")
	lenOfParts := len(parts)
	for index, part := range parts {
		if index == (lenOfParts - 1) {
			realValue = response[part]
			break
		}
		if response[part] == nil {
			break
		}
		response = response[part].(map[string]interface{})
	}

	return realValue
}

func extract(value string) interface{} {
	if v, err := strconv.Atoi(value); err == nil {
		return v
	}
	return value
}

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

//FetchActionType fetches action type
func FetchActionType(ruleVal string) string {
	parts := strings.Split(ruleVal, ".")
	if len(parts) > 2 {
		return parts[2]
	}
	return ""
}

//FetchActionKeyValue fetches action value to be updated
func FetchActionKeyValue(ruleVal string) (string, string) {
	parts := strings.Split(ruleVal, "=")
	if len(parts) > 1 {
		subParts := strings.Split(parts[0], ".")
		return subParts[len(subParts)-1], parts[1]
	}
	return "", ""
}

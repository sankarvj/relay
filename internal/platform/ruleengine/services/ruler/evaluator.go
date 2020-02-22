package ruler

import (
	"log"
	"reflect"
	"strconv"
	"strings"
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

func fetchRootKey(value string) string {
	parts := strings.Split(value, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func extract(value string) interface{} {
	if v, err := strconv.Atoi(value); err == nil {
		return v
	}
	return value
}

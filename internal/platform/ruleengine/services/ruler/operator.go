package ruler

import (
	"fmt"
	"log"
	"reflect"

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
	switch v := left.(type) {
	default:
		fmt.Printf("unexpected type %T", v)
	case int:
	case string:
		if version.ValidSimpleVersionFormat(left.(string)) && version.ValidSimpleVersionFormat(right.(string)) {
			r := version.CompareSimple(left.(string), right.(string))
			if r == 0 {
				return true
			}
			return false
		}
	}

	return left == right
}

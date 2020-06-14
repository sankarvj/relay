package ruler

import (
	"fmt"
	"log"

	version "github.com/mcuadros/go-version"
)

func compare(left, right Operand) bool {
	log.Printf("compare left: %s(%T) vs right: %s(%T)", left, left, right, right)

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

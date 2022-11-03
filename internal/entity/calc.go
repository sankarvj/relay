package entity

import (
	"fmt"
)

//MetaKeyCalc
type Calculator interface {
	Calc(oldValue, newValue interface{}) interface{}
}

type Aggr struct {
}

type Sum struct {
}

type Unknown struct {
}

func pickCalculator(calType string) Calculator {
	switch calType {
	case MetaCalcAggr:
		return Aggr{}
	case MetaCalcSum:
		return Sum{}
	default:
		return Unknown{}
	}
}

func (s Sum) Calc(oldValue, newValue interface{}) interface{} {
	switch oldValue.(type) {
	default:
		return newValue
	case int:
		return oldValue.(int) + newValue.(int)
	case int64:
		return oldValue.(int64) + int64(newValue.(int64))
	case float64:
		return oldValue.(float64) + float64(newValue.(float64))
	case string:

		return fmt.Sprintf("%s %s", oldValue, newValue)
	case bool:
		return newValue
	case []interface{}:
		return append(oldValue.([]interface{}), newValue.([]interface{})...)
	}
}

func (s Aggr) Calc(oldValue, newValue interface{}) interface{} {
	switch oldValue.(type) {
	default:
		return newValue
	case int:
		return oldValue.(int) + newValue.(int)
	case int64:
		return oldValue.(int64) + newValue.(int64)
	case float64:
		return oldValue.(float64) + newValue.(float64)
	case string:
		return fmt.Sprintf("%s %s", oldValue, newValue)
	case bool:
		return newValue
	case []interface{}:
		return append(oldValue.([]interface{}), newValue.([]interface{})...)
	}
}

func (s Unknown) Calc(oldValue, newValue interface{}) interface{} {
	return newValue
}

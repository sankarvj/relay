package ruler

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	version "github.com/mcuadros/go-version"
)

//OperandDT defines the datatype of operand
type OperandDT int

//DataType enum
const (
	UnknownDT OperandDT = iota
	StrDT
	NumberDT
	VersionDT
	ListDT
)

type caster struct {
	leftNumber    float64
	rightNumber   float64
	leftString    string
	rightString   string
	leftDataType  OperandDT
	rightDataType OperandDT
	err           error
	casters       []interface{}
}

func Compare(left, right interface{}) bool {
	return compare(left, right)
}

func compare(left, right Operand) bool {

	c := cast(left, right, true)
	if c.err != nil {
		log.Println("eq error comparing operands", c.err)
		return false
	}

	switch c.leftDataType {
	case StrDT:
		return c.leftString == c.rightString
	case NumberDT:
		return c.leftNumber == c.rightNumber
	case VersionDT:
		if version.CompareSimple(c.leftString, c.rightString) == 0 {
			return true
		}
	}
	return false
}

func greaterThan(left, right Operand) bool {
	c := cast(left, right, true)
	if c.err != nil {
		log.Println("gt error comparing operands", c.err)
		return false
	}
	switch c.leftDataType {
	case NumberDT:
		return c.leftNumber > c.rightNumber
	}
	return false
}

func lesserThan(left, right Operand) bool {
	c := cast(left, right, true)
	if c.err != nil {
		log.Println("lt error comparing operands", c.err)
		return false
	}

	switch c.leftDataType {
	case NumberDT:
		return c.leftNumber < c.rightNumber
	}
	return false
}

func in(left, right Operand) bool {
	c := cast(left, right, false)
	if c.err != nil {
		log.Println("in error comparing operands", c.err)
		return false
	}
	switch c.leftDataType {
	case ListDT:
		switch c.rightDataType {
		case StrDT:
			for _, v := range c.casters {
				if compare(v, c.rightString) {
					return true
				}
			}
		case NumberDT:
			for _, v := range c.casters {
				if compare(v, c.rightNumber) {
					return true
				}
			}
		}
	}
	return false
}

func cast(left, right Operand, checkEquality bool) caster {
	log.Printf("compare left: %v (%T) vs right: %v (%T)", left, left, right, right)
	c := caster{}
	if left == nil || right == nil {
		c.err = errors.New("Any one or both the operands are null")
		return c
	}
	c.setLeft(left)
	c.setRight(right)

	if checkEquality && (c.rightDataType != c.leftDataType) {
		c.err = fmt.Errorf("Can't do operation in two different operand types %v & %v", c.leftDataType, c.rightDataType)
	}
	return c
}

func (c *caster) setLeft(left Operand) {
	switch v := left.(type) {
	default:
		c.err = fmt.Errorf("unexpected type %T", v)
	case int:
		c.leftNumber = float64(left.(int))
		c.leftDataType = NumberDT
	case int64:
		c.leftNumber = float64(left.(int64))
		c.leftDataType = NumberDT
	case float64:
		c.leftNumber = float64(left.(float64))
		c.leftDataType = NumberDT
	case string:
		c.leftString = left.(string)
		c.leftDataType = c.deepCaster()
	case bool:
		c.leftString = strconv.FormatBool(left.(bool))
		c.leftDataType = c.deepCaster()
	case []interface{}:
		c.casters = left.([]interface{})
		c.leftDataType = ListDT
	}
}

func (c *caster) setRight(right Operand) {
	if c.err != nil {
		return
	}
	switch v := right.(type) {
	default:
		c.err = fmt.Errorf("unexpected type %T", v)
	case int:
		c.rightNumber = float64(right.(int))
		c.rightDataType = NumberDT
	case int64:
		c.rightNumber = float64(right.(int64))
		c.rightDataType = NumberDT
	case float64:
		c.rightNumber = float64(right.(float64))
		c.rightDataType = NumberDT
	case string:
		c.rightString = right.(string)
		c.rightDataType = c.deepCaster()
	case bool:
		c.rightString = strconv.FormatBool(right.(bool))
		c.rightDataType = c.deepCaster()
	case []interface{}:
		c.casters = right.([]interface{})
		c.rightDataType = ListDT
	}
}

func (c *caster) deepCaster() OperandDT {
	if version.ValidSimpleVersionFormat(c.leftString) && version.ValidSimpleVersionFormat(c.rightString) {
		return VersionDT
	}

	f64L, errL := strconv.ParseFloat(c.leftString, 64)
	f64R, errR := strconv.ParseFloat(c.rightString, 64)
	if errL == nil && errR == nil {
		c.leftNumber = float64(f64L)
		c.rightNumber = float64(f64R)
		return NumberDT
	}

	i64L, errL := strconv.ParseInt(c.leftString, 10, 32)
	i64R, errR := strconv.ParseInt(c.rightString, 10, 32)
	if errL == nil && errR == nil {
		c.leftNumber = float64(i64L)
		c.rightNumber = float64(i64R)
		return NumberDT
	}

	return StrDT

}

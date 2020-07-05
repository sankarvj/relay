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
	StrDT OperandDT = iota
	NumberDT
	VersionDT
	UnknownDT
)

type caster struct {
	leftNumber  float64
	rightNumber float64
	leftString  string
	rightString string
	dataType    OperandDT
	err         error
}

func compare(left, right Operand) bool {
	c := cast(left, right)
	if c.err != nil {
		log.Println("error comparing operands", c.err)
		return false
	}

	switch c.dataType {
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
	c := cast(left, right)
	if c.err != nil {
		log.Println("error comparing operands", c.err)
		return false
	}
	switch c.dataType {
	case NumberDT:
		return c.leftNumber > c.rightNumber
	}
	return false
}

func lesserThan(left, right Operand) bool {
	c := cast(left, right)
	if c.err != nil {
		log.Println("error comparing operands", c.err)
		return false
	}

	switch c.dataType {
	case NumberDT:
		return c.leftNumber < c.rightNumber
	}
	return false
}

func cast(left, right Operand) caster {
	log.Printf("compare left: %v (%T) vs right: %v (%T)", left, left, right, right)
	c := caster{}
	if left == nil || right == nil {
		c.err = errors.New("Any one or both the operands are null")
		return c
	}
	c.setLeft(left)
	c.setRight(right)
	return c
}

func (c *caster) setLeft(left Operand) {
	switch v := left.(type) {
	default:
		c.err = fmt.Errorf("unexpected type %T", v)
	case int:
		c.leftNumber = float64(left.(int))
		c.dataType = NumberDT
	case int64:
		c.leftNumber = float64(left.(int64))
		c.dataType = NumberDT
	case float64:
		c.leftNumber = float64(left.(float64))
		c.dataType = NumberDT
	case string:
		c.leftString = left.(string)
		c.dataType = c.deepCaster()
	}
}

func (c *caster) setRight(right Operand) {
	if c.err != nil {
		return
	}
	rightDataType := UnknownDT
	switch v := right.(type) {
	default:
		c.err = fmt.Errorf("unexpected type %T", v)
	case int:
		c.rightNumber = float64(right.(int))
		rightDataType = NumberDT
	case int64:
		c.rightNumber = float64(right.(int64))
		rightDataType = NumberDT
	case float64:
		c.rightNumber = float64(right.(float64))
		rightDataType = NumberDT
	case string:
		c.rightString = right.(string)
		rightDataType = c.deepCaster()
	}
	if rightDataType != c.dataType {
		c.err = fmt.Errorf("Can't do operation in two different operand types %v & %v", c.dataType, rightDataType)
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

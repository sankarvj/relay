package ruler

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	version "github.com/mcuadros/go-version"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
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
	TimeDT
)

type caster struct {
	leftNumber         float64
	rightNumber        float64
	leftString         string
	rightString        string
	leftTime           time.Time
	rightTime          time.Time
	leftDataType       OperandDT
	rightDataType      OperandDT
	err                error
	leftInterfaceList  []interface{}
	rightInterfaceList []interface{}
}

func Compare(left, right interface{}) bool {
	return compare(left, right)
}

func compare(left, right Operand) bool {
	c := cast(left, right, true)
	if c.err != nil {
		log.Println("unexpected error occurred when comparing operands for operator: eq and error:", c.err)
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
	case ListDT:
		return same(c.leftInterfaceList, c.rightInterfaceList)
	}
	return false
}

func differ(left, right Operand) bool {
	c := cast(left, right, true)
	if c.err != nil {
		log.Println("unexpected error occurred when comparing operands for operator: !eq and error:", c.err)
		return false
	}

	switch c.leftDataType {
	case StrDT:
		return c.leftString != c.rightString
	case NumberDT:
		return c.leftNumber != c.rightNumber
	case VersionDT:
		if version.CompareSimple(c.leftString, c.rightString) != 0 {
			return true
		}
	case ListDT:
		return notsame(c.leftInterfaceList, c.rightInterfaceList)
	}
	return false
}

func greaterThan(left, right Operand) bool {
	c := cast(left, right, true)
	if c.err != nil {
		log.Println("unexpected error occurred when comparing operands for operator: gt and error:", c.err)
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
		log.Println("unexpected error occurred when comparing operands for operator: lt and error:", c.err)
		return false
	}

	switch c.leftDataType {
	case NumberDT:
		return c.leftNumber < c.rightNumber
	}
	return false
}

func after(left, right Operand) bool {
	c := cast(left, right, true)
	if c.err != nil {
		log.Println("unexpected error occurred when comparing operands for operator: af and error:", c.err)
		return false
	}

	switch c.leftDataType {
	case TimeDT:
		return c.leftTime.After(c.rightTime)
	}
	return false
}

func before(left, right Operand) bool {
	c := cast(left, right, true)
	if c.err != nil {
		log.Println("unexpected error occurred when comparing operands for operator: bf and error:", c.err)
		return false
	}
	switch c.leftDataType {
	case TimeDT:
		return c.leftTime.Before(c.rightTime)
	}
	return false
}

func between(left, right Operand) bool { // assumption: left is a interface list & right is a simple string/number
	c := cast(left, right, false)
	if c.err != nil {
		log.Println("unexpected error occurred when comparing operands for operator: bw and error:", c.err)
		return false
	}

	min := time.Now()
	max := time.Now()
	switch c.rightDataType {
	case VersionDT:
		parts := strings.Split(c.rightString, "-")
		if len(parts) == 2 {
			min = util.ConvertMillisToTime(parts[0])
			max = util.ConvertMillisToTime(parts[1])
		}
	}

	log.Println("c.min ", min)
	log.Println("c.max ", max)

	switch c.leftDataType {
	case TimeDT:
		return c.leftTime.After(min) && c.leftTime.Before(max)
	}
	return false
}

func in(left, right Operand) bool { // assumption: left is a interface list & right is a simple string/number
	c := cast(left, right, false)
	if c.err != nil {
		log.Println("unexpected error occurred when comparing operands for operator: in and error:", c.err)
		return false
	}
	switch c.leftDataType {
	case ListDT:
		for _, v := range c.leftInterfaceList {
			if compare(v, right) {
				return true
			}
		}
	}
	return false
}

func notin(left, right Operand) bool { // assumption: left is a interface list & right is a simple string/number
	c := cast(left, right, false)
	if c.err != nil {
		log.Println("unexpected error occurred when comparing operands for operator: not in and error:", c.err)
		return false
	}
	switch c.leftDataType {
	case ListDT:
		for _, v := range c.leftInterfaceList {
			if compare(v, right) {
				return false
			}
		}
	}
	return true
}

func like(left, right Operand) bool {
	c := cast(left, right, true)
	if c.err != nil {
		log.Println("unexpected error occurred when comparing operands for operator: lk and error:", c.err)
		return false
	}
	switch c.leftDataType {
	case StrDT:
		return strings.HasPrefix(c.leftString, c.rightString)
	}
	return false
}

func same(leftList, rightList []interface{}) bool {
	matched := false
	if len(leftList) != len(rightList) {
		return false
	} else if len(leftList) == 0 && len(rightList) == 0 {
		return true
	} else {
		for _, lv := range leftList { //sorted list optimise: use i,j looping instead of range
			matched = false
			for _, rv := range rightList { // optimise: move j wrt to i
				if compare(lv, rv) {
					matched = true
					break
				}
			}
			if !matched {
				return false
			}
		}
	}

	return matched
}

func notsame(leftList, rightList []interface{}) bool {
	notmatched := true
	if len(leftList) != len(rightList) {
		return notmatched
	}

	for _, lv := range leftList {
		notmatched = false
		for _, rv := range rightList {
			if !compare(lv, rv) {
				return true
			}
		}
	}
	return notmatched
}

func cast(left, right Operand, checkEquality bool) caster {
	log.Printf("internal.platform.ruleengine.services.ruler.operator : compare left: %v (%T) vs right: %v (%T)\n", left, left, right, right)
	c := caster{}
	if left == nil || right == nil {
		c.err = errors.New("Any one or both the operands are null")
		return c
	}
	c.setLeft(left)
	c.setRight(right)

	if checkEquality && (c.rightDataType != c.leftDataType) {
		c.err = fmt.Errorf("unexpected error occurred. can't do operation in two different operand types %v & %v\n", c.leftDataType, c.rightDataType)
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
		c.leftDataType = c.deepCaster(true)
	case bool:
		c.leftString = strconv.FormatBool(left.(bool))
		c.leftDataType = c.deepCaster(true)
	case []interface{}:
		c.leftInterfaceList = left.([]interface{})
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
		c.rightDataType = c.deepCaster(false)
	case bool:
		c.rightString = strconv.FormatBool(right.(bool))
		c.rightDataType = c.deepCaster(false)
	case []interface{}:
		c.rightInterfaceList = right.([]interface{})
		c.rightDataType = ListDT
	}
}

func (c *caster) deepCaster(left bool) OperandDT {

	if left {
		//version
		if version.ValidSimpleVersionFormat(c.leftString) {
			return VersionDT
		}
		//float
		f64L, errL := strconv.ParseFloat(c.leftString, 64)
		if errL == nil {
			c.leftNumber = float64(f64L)
			return NumberDT
		}
		//int
		i64L, errL := strconv.ParseInt(c.leftString, 10, 32)
		if errL == nil {
			c.leftNumber = float64(i64L)
			return NumberDT
		}

		//other format time
		tL, errL := time.Parse("2006-01-02 15:04:05 -07:00", c.leftString)
		if errL == nil {
			c.leftTime = tL
			c.rightString = time.Now().Format("2006-01-02 15:04:05 -07:00") // do we need to set it here?
			return TimeDT
		}
	} else {
		//version
		if version.ValidSimpleVersionFormat(c.rightString) {
			return VersionDT
		}
		//float
		f64R, errR := strconv.ParseFloat(c.rightString, 64)
		if errR == nil {
			c.rightNumber = float64(f64R)
			return NumberDT
		}
		//int
		i64R, errR := strconv.ParseInt(c.rightString, 10, 32)
		if errR == nil {
			c.rightNumber = float64(i64R)
			return NumberDT
		}
		//time
		if c.rightString == "now" {
			c.rightString = time.Now().Format("2006-01-02 15:04:05 -07:00")
		}
		tR, errR := time.Parse("2006-01-02 15:04:05 -07:00", c.rightString)
		if errR == nil {
			c.rightTime = tR
			return TimeDT
		}
	}
	return StrDT
}

func findDT(right Operand) OperandDT {
	c := caster{}
	c.setRight(right)
	if c.err != nil {
		log.Println("unexpected error occurred when comparing operands. error:", c.err)
		return UnknownDT
	}
	return c.rightDataType
}

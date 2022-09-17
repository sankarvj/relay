package ruler

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

type Filter struct {
	Conditions map[string]Condition `json:"conditions"` // key is the field key
}

type Condition struct {
	Term       interface{} `json:"term"`
	Expression string      `json:"expression"`
	DataType   DType       `json:"d_type"`
}

//ExpressionType returns the type of expression
type ExpressionType int

// ExpressionType defines different types of expression to work with
const (
	Worker ExpressionType = iota
	Grapher
	Querier
	Parser
	Computer
	NegExecutor
	PosExecutor
)

//Work used to evaluate the expression
type Work struct {
	Type          ExpressionType
	Expression    string
	OutboundResp  interface{}
	InboundRespCh chan interface{}
}

//Ruler has the full set of rule items
type Ruler struct {
	RuleItem *RuleItem
	positive *bool
	trigger  string
	content  interface{}
	filter   *Filter
	//channels to send the results back and forth
	workChan chan Work
}

//RuleItem specifies the single section of the rules
type RuleItem struct {
	left     Operand
	right    Operand
	dataType DType

	//operation and the snippet
	isANDOp   bool
	operation applyFn
	operator  string
}

//DType defines the data type of field
type DType string

//Mode for the entity spcifies certain entity specific characteristics
//Keep this as minimal and add a sub-type for data types such as decimal,boolean,time & date
const (
	TypeUnKnown   DType = "U"
	TypeString          = "S"
	TypeNumber          = "N"
	TypeDataTime        = "T"
	TypeList            = "L"
	TypeReference       = "R"
)

//EngineFeedback defines the type of response execution
type EngineFeedback int

const (
	Execute EngineFeedback = iota
	Parse
	Graph
	Compute
)

//applyFn takes two operands and then apply the logic to get the final result
type applyFn func(Operand, Operand) bool

//Operand is the value of type interface
type Operand interface{}

//Run starts the lexer by passing the rule and a res chan,
//res chan will trigger when the rule engine needs a response in the form of map
func Run(rule string, eFeedback EngineFeedback, workChan chan Work) {
	defer close(workChan)

	if strings.TrimSpace(rule) == "" {
		//log.Println("internal.platform.ruleengine.services.ruler : encountered empty expression, sending positive response")
		// By default, the empty rule is considered as the positive expression.
		// stand taken since the default nodes don't possess expressions
		workChan <- Work{PosExecutor, "", nil, nil}
		return
	}
	r := Ruler{
		workChan: workChan,
		positive: nil, //always start on a nil note! :)
	}

	switch eFeedback {
	case Execute:
		//log.Printf("internal.platform.ruleengine.services.ruler case: `execute`  expression: %s\n", rule)
		r = r.startExecutingLexer(rule)
		if r.positive != nil && *r.positive {
			workChan <- Work{PosExecutor, r.trigger, nil, nil}
		} else {
			workChan <- Work{NegExecutor, r.trigger, nil, nil}
		}
	case Parse:
		//log.Printf("internal.platform.ruleengine.services.ruler case: `parse` expression: %s\n", rule)
		r = r.startParsingLexer(rule)
		//CHECK: This might cause adverse effects in the html contents. Take note
		workChan <- Work{Parser, "", r.content, nil}
	case Compute:
		//log.Printf("internal.platform.ruleengine.services.ruler case: `compute` expression: %s\n", rule)
		r = r.startComputingLexer(rule)
		workChan <- Work{Computer, "", r.content, nil}
	case Graph:
		//log.Printf("internal.platform.ruleengine.services.ruler case: `graph` expression: %s\n", rule)
		r = r.startGraphingLexer(rule)
		workChan <- Work{Grapher, "", r.filter, nil}
	}

}

func (r Ruler) startExecutingLexer(rule string) Ruler {
	l := lexer.BeginLexing("rule", rule)
	var token lexertoken.Token
	for {
		token = l.NextToken()
		switch token.Type {
		case lexertoken.TokenValuate:
			r.addEvalOperand(strings.TrimSpace(token.Value))
		case lexertoken.TokenEqualSign:
			r.addCompareOperation(false)
		case lexertoken.TokenNotEqualSign:
			r.addCompareOperation(true)
		case lexertoken.TokenGTSign:
			r.addGTCompareOperation()
		case lexertoken.TokenLTSign:
			r.addLTCompareOperation()
		case lexertoken.TokenONSign:
			r.addONCompareOperation()
		case lexertoken.TokenAFSign:
			r.addAFCompareOperation()
		case lexertoken.TokenBFSign:
			r.addBFCompareOperation()
		case lexertoken.TokenBWSign:
			r.addBWCompareOperation()
		case lexertoken.TokenINSign:
			r.addINOperation(false)
		case lexertoken.TokenNotINSign:
			r.addINOperation(true)
		case lexertoken.TokenANDOperation:
			r.addANDCondition()
		case lexertoken.TokenOROperation:
			r.addORCondition()
		case lexertoken.TokenValue:
			r.addOperand(extract(token.Value))
		case lexertoken.TokenSnippet:
			r.addTrigger(token.Value)
		case lexertoken.TokenRightBrace, lexertoken.TokenRightDoubleBrace:
			r.execute()
		case lexertoken.TokenGibberish:
			r.addGibbrish(token.Value)
		case lexertoken.TokenEOF:
			return r
		}
	}
}

func (r Ruler) startParsingLexer(rule string) Ruler {
	l := lexer.BeginLexing("rule", rule)
	var token lexertoken.Token
	for {
		token = l.NextToken()
		switch token.Type {
		case lexertoken.TokenValuate:
			r.addEvalOperand(strings.TrimSpace(token.Value))
		case lexertoken.TokenValue:
			r.addOperand(extract(token.Value))
		case lexertoken.TokenGibberish:
			r.addGibbrish(token.Value)
		case lexertoken.TokenEOF:
			return r
		}
	}
}

func (r Ruler) startComputingLexer(rule string) Ruler {
	log.Println("before  ReplaceHTML--:: ", rule)
	rule = ReplaceHTML(rule)
	log.Println("after ReplaceHTML--:: ", rule)
	l := lexer.BeginLexing("rule", rule)
	var token lexertoken.Token
	for {
		token = l.NextToken()
		switch token.Type {
		case lexertoken.TokenValuate:
			r.addEvalOperand(strings.TrimSpace(token.Value))
		case lexertoken.TokenValue:
			r.addOperand(extract(token.Value))
		case lexertoken.TokenGibberish:
			r.addGibbrish(token.Value)
		case lexertoken.TokenQuery: //special type to render the template info
			r.addQuery(strings.TrimSpace(token.Value))
		case lexertoken.TokenEOF:
			return r
		}
	}

}

func (r Ruler) startGraphingLexer(rule string) Ruler {
	l := lexer.BeginLexing("rule", rule)
	var token lexertoken.Token
	for {
		token = l.NextToken()
		switch token.Type {
		case lexertoken.TokenValuate:
			r.addEvalOperand(strings.TrimSpace(token.Value))
		case lexertoken.TokenEqualSign:
			r.addCompareOperation(false)
		case lexertoken.TokenNotEqualSign:
			r.addCompareOperation(true)
		case lexertoken.TokenGTSign:
			r.addGTCompareOperation()
		case lexertoken.TokenLTSign:
			r.addLTCompareOperation()
		case lexertoken.TokenONSign:
			r.addONCompareOperation()
		case lexertoken.TokenAFSign:
			r.addAFCompareOperation()
		case lexertoken.TokenBFSign:
			r.addBFCompareOperation()
		case lexertoken.TokenBWSign:
			r.addBWCompareOperation()
		case lexertoken.TokenINSign:
			r.addINOperation(false)
		case lexertoken.TokenNotINSign:
			r.addINOperation(true)
		case lexertoken.TokenLKSign:
			r.addLKOperation()
		case lexertoken.TokenANDOperation:
			r.addANDCondition()
		case lexertoken.TokenOROperation:
			r.addORCondition()
		case lexertoken.TokenValue:
			r.addOperand(extract(token.Value))
		case lexertoken.TokenGibberish:
			r.addGibbrish(token.Value)
		case lexertoken.TokenRightBrace, lexertoken.TokenRightDoubleBrace:
			r.makeGraph()
		case lexertoken.TokenEOF:
			return r
		}
	}
}

func (r *Ruler) addEvalOperand(value interface{}) error {
	return r.addOperand(r.eval(value.(string)))
}

func (r *Ruler) addGibbrish(value interface{}) error {
	r.setContent(value)
	return nil
}

func (r *Ruler) addOperand(value interface{}) error {
	//never set the value to nil. That will make the execute condition fail for valid cases
	if value == nil {
		value = "nil"
	}
	r.setContent(value)
	r.constructRuleItem()
	if r.RuleItem.left == nil {
		r.RuleItem.left = value
	} else if r.RuleItem.right == nil {
		r.RuleItem.right = value
	}
	return nil
}

func (r *Ruler) addCompareOperation(not bool) error {
	r.constructRuleItem()
	if not {
		r.RuleItem.operation = differ
		r.RuleItem.operator = lexertoken.NotEqualSign
	} else {
		r.RuleItem.operation = compare
		r.RuleItem.operator = lexertoken.EqualSign
	}

	return nil
}

func (r *Ruler) addGTCompareOperation() error {
	r.constructRuleItem()
	r.RuleItem.operation = greaterThan
	r.RuleItem.operator = lexertoken.GTSign
	return nil
}

func (r *Ruler) addLTCompareOperation() error {
	r.constructRuleItem()
	r.RuleItem.operation = lesserThan
	r.RuleItem.operator = lexertoken.LTSign
	return nil
}

func (r *Ruler) addLKOperation() error {
	r.constructRuleItem()
	r.RuleItem.operation = like
	r.RuleItem.operator = lexertoken.LikeSign
	return nil
}

func (r *Ruler) addONCompareOperation() error { // TODO. Implementation missing
	r.constructRuleItem()
	r.RuleItem.operation = between
	r.RuleItem.operator = lexertoken.BWSign
	return nil
}

func (r *Ruler) addAFCompareOperation() error {
	r.constructRuleItem()
	r.RuleItem.operation = after
	r.RuleItem.operator = lexertoken.AFSign
	return nil
}

func (r *Ruler) addBFCompareOperation() error {
	r.constructRuleItem()
	r.RuleItem.operation = before
	r.RuleItem.operator = lexertoken.BFSign
	return nil
}

func (r *Ruler) addBWCompareOperation() error { // TODO. Implementation missing
	r.constructRuleItem()
	r.RuleItem.operation = between
	r.RuleItem.operator = lexertoken.BWSign
	return nil
}

func (r *Ruler) addINOperation(not bool) error {
	r.constructRuleItem()
	if not {
		r.RuleItem.operation = notin
		r.RuleItem.operator = lexertoken.NotINSign
	} else {
		r.RuleItem.operation = in
		r.RuleItem.operator = lexertoken.INSign
	}
	return nil
}

func (r *Ruler) addANDCondition() error {
	r.constructRuleItem()
	r.RuleItem.isANDOp = true
	return nil
}

func (r *Ruler) addORCondition() error {
	r.constructRuleItem()
	r.RuleItem.isANDOp = false
	return nil
}

func (r *Ruler) addTrigger(trigger string) error {
	r.trigger = trigger
	return nil
}

func (r *Ruler) constructRuleItem() {
	if r.RuleItem == nil {
		r.RuleItem = &RuleItem{}
	}
}

func (r *Ruler) saveAndResetRuleItem(singleUnitResult bool) {
	temp := false
	if r.RuleItem.isANDOp {
		if r.positive == nil {
			r.positive = newTrue() // start it with true for AND
		}
		temp = *r.positive && singleUnitResult
	} else { // any condition ---> OR case
		if r.positive == nil {
			r.positive = newFalse() // start it with false for OR
		}
		temp = *r.positive || singleUnitResult
	}
	r.positive = &temp
	r.RuleItem = nil
}

func (r *Ruler) addQuery(q string) {
	respChan := make(chan interface{})
	r.workChan <- Work{Querier, q, nil, respChan}
	resp := <-respChan
	r.setContent(resp)
	r.constructRuleItem()
	r.saveAndResetRuleItem(true)
}

func (r *Ruler) execute() error {
	r.constructRuleItem()
	//DEBUG LOGS
	log.Printf("*********> debug: internal.platform.ruleengine.services.ruler : `execute:` left_rule_item: %v | right_rule_item: %v | op: %+v | isAND: %t\n", r.RuleItem.left, r.RuleItem.right, r.RuleItem.operation, r.RuleItem.isANDOp)

	var opResult bool
	if r.RuleItem.left == "nil" && r.RuleItem.right == "nil" {
		opResult = false
	} else if r.RuleItem.left != nil && r.RuleItem.right != nil && r.RuleItem.operation != nil {
		opResult = r.RuleItem.operation(r.RuleItem.left, r.RuleItem.right)
	} else {
		//Its in the middle. Don't execute
		return nil
	}
	//save the result of the unit and reset the ruleItem
	r.saveAndResetRuleItem(opResult)
	return nil
}

func (r *Ruler) makeGraph() error {
	if r.filter == nil {
		r.filter = &Filter{
			Conditions: map[string]Condition{},
		}
	}
	r.constructRuleItem()
	//log.Printf("internal.platform.ruleengine.services.ruler : `query:` execute left_rule_item: %v | right_rule_item: %v | op: %+v | isAND: %t\n", r.RuleItem.left, r.RuleItem.right, r.RuleItem.operation, r.RuleItem.isANDOp)

	if r.RuleItem.left != nil && r.RuleItem.right != nil && r.RuleItem.operation != nil {
		condition := Condition{
			Expression: r.RuleItem.operator,
			DataType:   dtype(findDT(r.RuleItem.right)),
			Term:       r.RuleItem.right,
		}
		key := r.RuleItem.left.(string)
		r.filter.Conditions[key] = condition
	} else {
		//Its in the middle. Don't execute
		return nil
	}

	//save the result of the unit and reset the ruleItem
	r.saveAndResetRuleItem(true)
	return nil
}

func (r *Ruler) exit() error {
	log.Println("*> internal.platforms.ruleengine.services.ruler exit")
	return nil
}

func (r *Ruler) eval(expression string) interface{} {
	respChan := make(chan interface{})
	r.workChan <- Work{Worker, expression, nil, respChan}
	resp := <-respChan
	return resp
}

func (r *Ruler) setContent(value interface{}) {
	if value == nil || value == "" {
		return
	}
	if r.content != nil { //parser will go here....
		//In go strings are immutable. Hence, this line is inefficient.
		r.content = fmt.Sprintf("%s %v", r.content, value)
	} else { // computer will go here. thus keeping the value as interface for the computer parser.
		r.content = value
	}

}

func newTrue() *bool {
	a := true
	return &a
}

func newFalse() *bool {
	b := false
	return &b
}

func extract(value string) interface{} {
	if v, err := strconv.Atoi(value); err == nil {
		return v
	}
	return value
}

func dtype(opDT OperandDT) DType {
	switch opDT {
	case NumberDT:
		return TypeNumber
	case StrDT:
		return TypeString
	case VersionDT:
		return TypeString
	case ListDT:
		return TypeList
	case TimeDT:
		return TypeDataTime
	case UnknownDT:
		return TypeUnKnown
	default:
		return TypeUnKnown
	}
}

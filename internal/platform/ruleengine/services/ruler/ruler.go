package ruler

import (
	"fmt"
	"log"
	"strings"

	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

//ExpressionType returns the type of expression
type ExpressionType int

// ExpressionType defines different types of expression to work with
const (
	Worker ExpressionType = iota
	NegExecutor
	PosExecutor
	Content
)

//Work used to evaluate the expression
type Work struct {
	Type       ExpressionType
	Expression string
	Resp       chan map[string]interface{}
}

//Ruler has the full set of rule items
type Ruler struct {
	RuleItem *RuleItem
	positive *bool
	trigger  string
	content  string
	//channels to send the results back and forth
	workChan chan Work
}

//RuleItem specifies the single section of the rules
type RuleItem struct {
	left  Operand
	right Operand

	//operation and the snippet
	isANDOp   bool
	operation applyFn
}

//applyFn takes two operands and then apply the logic to get the final result
type applyFn func(Operand, Operand) bool

//Operand is the value of type interface
type Operand interface{}

//Run starts the lexer by passing the rule and a res chan,
//res chan will trigger when the rule engine needs a response in the form of map
func Run(rule string, isExecutor bool, workChan chan Work) {
	defer close(workChan)
	if rule == "" {
		// by default the empty rule is considered as the positive expression. Since most of nodes don't possess expressions
		workChan <- Work{PosExecutor, "", nil}
		return
	}
	log.Println("Starting lexer and parser for rule - ", rule, "...")
	r := Ruler{
		workChan: workChan,
		positive: nil, //always start on a nil note! :)
	}
	if isExecutor {
		r = r.startExecutingLexer(rule)
		if r.positive != nil && *r.positive {
			workChan <- Work{PosExecutor, r.trigger, nil}
		} else {
			workChan <- Work{NegExecutor, r.trigger, nil}
		}
	} else {
		r = r.startParsingLexer(rule)
		//NOTE: This might cause adverse effects in the html contents. Take note
		r.content = strings.TrimSpace(r.content)
		workChan <- Work{Content, r.content, nil}
	}

}

func (r Ruler) startExecutingLexer(rule string) Ruler {
	l := lexer.BeginLexing("rule", rule)
	var token lexertoken.Token
	for {
		token = l.NextToken()
		//log.Println("token", token)
		switch token.Type {
		case lexertoken.TokenValuate:
			r.addEvalOperand(strings.TrimSpace(token.Value))
		case lexertoken.TokenEqualSign:
			r.addCompareOperation()
		case lexertoken.TokenGTSign:
			r.addGTCompareOperation()
		case lexertoken.TokenLTSign:
			r.addLTCompareOperation()
		case lexertoken.TokenINSign:
			r.addINOperation()
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
		case lexertoken.TokenQuery:
			r.query(strings.TrimSpace(token.Value))
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

func (r *Ruler) addCompareOperation() error {
	r.constructRuleItem()
	r.RuleItem.operation = compare
	return nil
}

func (r *Ruler) addGTCompareOperation() error {
	r.constructRuleItem()
	r.RuleItem.operation = greaterThan
	return nil
}

func (r *Ruler) addLTCompareOperation() error {
	r.constructRuleItem()
	r.RuleItem.operation = lesserThan
	return nil
}

func (r *Ruler) addINOperation() error {
	r.constructRuleItem()
	r.RuleItem.operation = in
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

func (r *Ruler) query(q string) {
	log.Printf("query %v", q)
	r.constructRuleItem()
	opResult := true
	r.saveAndResetRuleItem(opResult)
}

func (r *Ruler) execute() error {
	r.constructRuleItem()
	log.Printf("execute left_rule_item: %v | right_rule_item: %v | op: %+v | isAND: %t", r.RuleItem.left, r.RuleItem.right, r.RuleItem.operation, r.RuleItem.isANDOp)

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

func (r *Ruler) exit() error {
	log.Println("exit exit exit exit exit exit exit exit")
	return nil
}

func (r *Ruler) eval(expression string) interface{} {
	respChan := make(chan map[string]interface{})
	r.workChan <- Work{Worker, expression, respChan}
	resp := <-respChan
	return Evaluate(expression, resp)
}

func (r *Ruler) setContent(value interface{}) {
	//In go strings are immutable. Hence, this line is inefficient.
	r.content = fmt.Sprintf("%s %v", r.content, value)
}

func newTrue() *bool {
	a := true
	return &a
}

func newFalse() *bool {
	b := false
	return &b
}

func join(strs ...string) string {
	var sb strings.Builder
	for _, str := range strs {
		sb.WriteString(str)
	}
	return sb.String()
}

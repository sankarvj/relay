package ruler

import (
	"log"
	"strings"

	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

//Work used to evaluate the expression
type Work struct {
	Expression string
	Resp       chan map[string]interface{}
}

//Ruler has the full set of rule items
type Ruler struct {
	RuleItem *RuleItem
	positive bool
	trigger  string
	//channels to send the results back
	workChan    chan Work
	triggerChan chan string
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
func Run(rule string, workChan chan Work, triggerChan chan string) {
	log.Println("Starting lexer and parser for rule - ", rule, "...")
	r := Ruler{
		workChan:    workChan,
		triggerChan: triggerChan,
		positive:    true, //always start on a positive note! :)
	}
	r = r.startLexer(rule)
	if r.positive {
		r.triggerChan <- r.trigger
	}
	close(r.triggerChan)
	close(r.workChan)
}

func (r Ruler) startLexer(rule string) Ruler {
	l := lexer.BeginLexing("rule", rule)
	var token lexertoken.Token
	for {
		token = l.NextToken()
		log.Println("token", token)
		switch token.Type {
		case lexertoken.TokenValuate:
			r.addEvalOperand(strings.TrimSpace(token.Value))
		case lexertoken.TokenEqualSign:
			r.addCompareOperation()
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
		case lexertoken.TokenRightSnippet:
			r.exit()
			return r
		case lexertoken.TokenEOF:
			return r
		}
	}
}

func (r *Ruler) addEvalOperand(value interface{}) error {
	return r.addOperand(r.eval(value.(string)))
}

func (r *Ruler) addOperand(value interface{}) error {
	//never set the value to nil. That will make the execute condition fail for valid cases
	if value == nil {
		value = "nil"
	}

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
	if r.RuleItem.isANDOp {
		r.positive = r.positive && singleUnitResult
	} else { // any condition ---> OR case
		r.positive = r.positive || singleUnitResult
	}
	r.RuleItem = nil
}

func (r *Ruler) execute() error {
	r.constructRuleItem()
	log.Println("execute r.RuleItem.left ", r.RuleItem.left)
	log.Println("execute r.RuleItem.right ", r.RuleItem.right)
	log.Println("execute r.RuleItem.operation ", r.RuleItem.operation)
	log.Println("execute r.RuleItem.isANDOp ", r.RuleItem.isANDOp)

	var opResult bool
	if r.RuleItem.left == "nil" && r.RuleItem.right == "nil" {
		opResult = false
	} else if r.RuleItem.left != nil && r.RuleItem.right != nil && r.RuleItem.operation != nil {
		opResult = r.RuleItem.operation(r.RuleItem.left, r.RuleItem.right)
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
	r.workChan <- Work{expression, respChan}
	resp := <-respChan
	return evaluate(expression, resp)
}
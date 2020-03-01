package ruler

import (
	"encoding/json"
	"errors"
	"log"
	"strings"

	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

//Work used to evaluate the expression
type Work struct {
	Key         string
	CurrentRule string
	Resp        chan map[string]interface{}
}

//Ruler has the full set of rule items
type Ruler struct {
	RuleItems []RuleItem
	itemCount int
	//channels to send the results back
	action chan ActionItem
	work   chan Work
}

//RuleItem specifies the single section of the rules
type RuleItem struct {
	left  Operand
	right Operand
	//operation and the snippet
	operation  applyFn
	actionItem ActionItem
}

// ActionItem specifies the action
type ActionItem struct {
	Set         map[string]interface{}
	Condition   map[string]interface{}
	Uncondition map[string]interface{}
}

//applyFn takes two operands and then apply the logic to get the final result
type applyFn func(Operand, Operand) bool

//Operand is the value of type interface
type Operand interface{}

//Run starts the lexer by passing the rule and a res chan,
//res chan will trigger when the rule engine needs a response in the form of map
func Run(rule string, work chan Work, actionItem chan ActionItem) {
	log.Println("Starting lexer and parser for rule - ", rule, "...")
	r := Ruler{
		action: actionItem,
		work:   work,
	}
	r = r.build(rule)
	for index := 0; index < len(r.RuleItems); index++ {
		item := r.RuleItems[index]
		positive := item.operation(item.left, item.right)
		if positive {
			r.action <- item.actionItem
		}
	}
	close(r.action)
	close(r.work)
}

func (r Ruler) build(rule string) Ruler {
	l := lexer.BeginLexing("rule", rule)
	var token lexertoken.Token
	for {
		token = l.NextToken()
		log.Println("token", token)
		switch token.Type {
		case lexertoken.TokenValuate:
			r.addOperand(strings.TrimSpace(token.Value), true)
		case lexertoken.TokenEqualSign:
			r.addCompareOperation()
		case lexertoken.TokenValue:
			r.addOperand(extract(token.Value), false)
		case lexertoken.TokenSnippet:
			r.addActionSnippet(token.Value)
		case lexertoken.TokenEOF:
			return r
		}
	}
}

func (r *Ruler) addOperand(value interface{}, eval bool) error {
	if eval {
		value = r.eval(value.(string))
	}

	if !r.isInMiddleOfRule() {
		r.RuleItems = append(r.RuleItems, RuleItem{left: value})
		r.itemCount = r.itemCount + 1
	} else {
		r.RuleItems[r.itemCount-1].right = value
	}
	return nil
}

func (r *Ruler) addCompareOperation() error {
	if r.isInMiddleOfRule() {
		r.RuleItems[r.itemCount-1].operation = compare
		return nil
	}
	return errors.New("incorrect rule syntax")
}

func (r *Ruler) isInMiddleOfRule() bool {
	return r.itemCount%2 != 0
}

func (r *Ruler) addActionSnippet(actionSnippet string) error {
	var actionItem ActionItem
	if err := json.Unmarshal([]byte(actionSnippet), &actionItem); err != nil {
		return err
	}
	for key, val := range actionItem.Set {
		actionItem.Set[key] = r.evalate(val.(string))
	}

	for key, val := range actionItem.Condition {
		actionItem.Condition[key] = r.evalate(val.(string))
	}

	for key, val := range actionItem.Uncondition {
		actionItem.Uncondition[key] = r.evalate(val.(string))
	}

	r.RuleItems[r.itemCount-1].actionItem = actionItem
	return nil
}

func (r Ruler) evalate(val string) interface{} {
	if strings.HasPrefix(val, lexertoken.LeftDoubleBraces) {
		val = val[len(lexertoken.LeftDoubleBraces):(len(val) - len(lexertoken.RightDoubleBraces))]
		return r.eval(val)
	}
	return val
}

func (r *Ruler) eval(val string) interface{} {
	key := FetchRootKey(val)
	respChan := make(chan map[string]interface{})
	r.work <- Work{key, val, respChan}
	resp := <-respChan
	return evaluate(val, resp)
}

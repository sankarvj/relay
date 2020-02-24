package ruler

import (
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

	action chan string
	work   chan Work
}

//RuleItem specifies the single section of the rules
type RuleItem struct {
	left  Operand
	right Operand

	operation     applyFn
	actionSnippet string
}

//applyFn takes two operands and then apply the logic to get the final result
type applyFn func(Operand, Operand) bool

//Operand is the value of type interface
type Operand interface{}

//Run starts the lexer by passing the rule and a res chan,
//res chan will trigger when the rule engine needs a response in the form of map
func Run(rule string, work chan Work, action chan string) {
	log.Println("Starting lexer and parser for rule - ", rule, "...")
	r := Ruler{
		action: action,
		work:   work,
	}
	r = r.build(rule)
	for index := 0; index < len(r.RuleItems); index++ {
		item := r.RuleItems[index]
		positive := item.operation(item.left, item.right)
		if positive {
			r.action <- item.actionSnippet
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
		key := FetchRootKey(value.(string))
		respChan := make(chan map[string]interface{})
		r.work <- Work{key, value.(string), respChan}
		resp := <-respChan
		value = evaluate(value.(string), resp)
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

func (r *Ruler) addActionSnippet(value string) error {
	r.RuleItems[r.itemCount-1].actionSnippet = value
	return nil
}

func (r *Ruler) isInMiddleOfRule() bool {
	return r.itemCount%2 != 0
}

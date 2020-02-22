package ruler

import (
	"errors"
	"log"
	"strings"

	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
)

//Ruler has the full set of rule items
type Ruler struct {
	RuleItems []RuleItem
	itemCount int

	action chan string
	worker workerFn
}

//RuleItem specifies the single section of the rules
type RuleItem struct {
	left  Operand
	right Operand

	operation     applyFn
	actionSnippet string
}

//hashKeyFn takes the keys and returns the response json in the form of map
type workerFn func(string) map[string]interface{}

//applyFn takes two operands and then apply the logic to get the final result
type applyFn func(Operand, Operand) bool

//Operand is the value of type interface
type Operand interface{}

//Run starts the lexer by passing the rule and a res chan,
//res chan will trigger when the rule engine needs a response in the form of map
func Run(rule string, worker workerFn, action chan string) {
	log.Println("Starting lexer and parser for rule - ", rule, "...")
	ruler := Ruler{
		action: action,
		worker: worker,
	}
	ruler = ruler.build(rule)
	for index := 0; index < len(ruler.RuleItems); index++ {
		item := ruler.RuleItems[index]
		positive := item.operation(item.left, item.right)
		if positive {
			ruler.action <- item.actionSnippet
		}
	}
	close(ruler.action)
}

func (ruler Ruler) build(rule string) Ruler {
	l := lexer.BeginLexing("rule", rule)
	var token lexertoken.Token
	for {
		token = l.NextToken()
		log.Println("token", token)
		switch token.Type {
		case lexertoken.TokenValuate:
			ruler.addOperand(strings.TrimSpace(token.Value), true)
		case lexertoken.TokenEqualSign:
			ruler.addCompareOperation()
		case lexertoken.TokenValue:
			ruler.addOperand(extract(token.Value), false)
		case lexertoken.TokenSnippet:
			ruler.addActionSnippet(token.Value)
		case lexertoken.TokenEOF:
			return ruler
		}
	}
}

func (r *Ruler) addOperand(value interface{}, eval bool) error {
	if eval {
		key := fetchRootKey(value.(string))
		value = evaluate(value.(string), r.worker(key))
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

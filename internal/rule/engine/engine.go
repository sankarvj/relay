package engine

import (
	"context"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/ruler"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

//GlobalEntity is the generic entity-id for certain expressions. See worker for its usecases
const (
	GlobalEntity = "xyz"
)

//RunRuleEngine runs the expression and execute action if the expression conditions met
func RunRuleEngine(ctx context.Context, db *sqlx.DB, n node.Node) (map[string]interface{}, error) {
	var err error
	var engineResponse map[string]interface{}
	signalsChan := make(chan ruler.Work)
	go ruler.Run(n.Expression, true, signalsChan)
	//signalsChan wait to receive evaluation work and final execution
	for work := range signalsChan {
		switch work.Type {
		case ruler.Worker:
			work.Resp <- worker(ctx, db, work.Expression, n.VariablesMap())
		case ruler.PosExecutor:
			engineResponse, err = executePosCase(ctx, db, n)
		case ruler.NegExecutor:
			engineResponse, err = executeNegCase(ctx, db, n)
		}
	}
	return engineResponse, err
}

//RunExpRenderer run the expression and returns evaluated string
func RunExpRenderer(ctx context.Context, db *sqlx.DB, exp string, variables map[string]string) string {
	var lexedContent string
	signalsChan := make(chan ruler.Work)
	go ruler.Run(exp, false, signalsChan)
	//signalsChan wait to receive evaluation work and final evaluated string
	for work := range signalsChan {
		switch work.Type {
		case ruler.Worker:
			work.Resp <- worker(ctx, db, work.Expression, variables)
		case ruler.Content:
			lexedContent = work.Expression
		}
	}
	return lexedContent
}

//RunExpEvaluator runs the expression to see whether the condition met or not
func RunExpEvaluator(ctx context.Context, db *sqlx.DB, exp string, variables map[string]string) bool {
	positive := false
	signalsChan := make(chan ruler.Work)
	go ruler.Run(exp, true, signalsChan)
	//signalsChan wait to receive evaluation work and final execution
	for work := range signalsChan {
		switch work.Type {
		case ruler.Worker:
			work.Resp <- worker(ctx, db, work.Expression, variables)
		case ruler.PosExecutor:
			positive = true
		}
	}
	return positive
}

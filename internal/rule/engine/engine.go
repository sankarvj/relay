package engine

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/jobber"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/ruler"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

type Engine struct {
	Job jobber.Jobber
}

// RuleResult returns back the recently executed results
type RuleResult struct {
	Executed bool
	Pause    bool
	Response map[string]interface{}
}

// RunRuleEngine runs the expression and execute action if the expression conditions met
func (e *Engine) RunRuleEngine(ctx context.Context, db *sqlx.DB, sdb *database.SecDB, n node.Node) (*RuleResult, error) {
	var err error
	signalsChan := make(chan ruler.Work)
	go ruler.Run(n.Expression, ruler.Execute, signalsChan)
	ruleResult := &RuleResult{}
	//signalsChan wait to receive evaluation work and final execution
	for work := range signalsChan {
		switch work.Type {
		case ruler.Worker:
			if result, err := worker(ctx, db, sdb, n.AccountID, work.Expression, n.VariablesMap()); err != nil {
				return nil, err
			} else {
				work.InboundRespCh <- result
			}
		case ruler.Grapher:
			if result, err := grapher(ctx, db, sdb, n.AccountID, work.Expression); err != nil {
				return nil, err
			} else {
				work.InboundRespCh <- result
			}
		case ruler.PosExecutor:
			err = ruleResult.executePosCase(ctx, e, n, db, sdb)
		case ruler.NegExecutor:
			err = ruleResult.executeNegCase(ctx, e, n, db, sdb)
		}
	}

	return ruleResult, err
}

// RunExpRenderer run the expression and returns evaluated string
func (e *Engine) RunExpRenderer(ctx context.Context, db *sqlx.DB, sdb *database.SecDB, accountID, exp string, variables map[string]interface{}) string {
	var lexedContent string
	signalsChan := make(chan ruler.Work)
	go ruler.Run(exp, ruler.Parse, signalsChan)
	//signalsChan wait to receive evaluation work and final evaluated string
	for work := range signalsChan {
		switch work.Type {
		case ruler.Worker: //input:executes this when it finds the {{}}
			if result, err := worker(ctx, db, sdb, accountID, work.Expression, variables); err != nil {
				return err.Error()
			} else {
				work.InboundRespCh <- result
			}
		case ruler.Parser: //output:executes this when it finds EOF/Response
			if work.OutboundResp != nil {
				lexedContent = work.OutboundResp.(string)
			} else {
				lexedContent = ""
			}

		}
	}
	return lexedContent
}

// RunFieldExpRenderer is same as RunExpRenderer but it evaluate the single value and return
// Added this new func to handle to evalution of expressions for the reference field which returns array of string. Instead of string
func (e *Engine) RunFieldExpRenderer(ctx context.Context, db *sqlx.DB, sdb *database.SecDB, accountID, exp string, variables map[string]interface{}) interface{} {
	var lexedContent interface{}
	signalsChan := make(chan ruler.Work)
	go ruler.Run(exp, ruler.Compute, signalsChan)
	//signalsChan wait to receive evaluation work and final evaluated string
	for work := range signalsChan {
		switch work.Type {
		case ruler.Worker: //input:executes this when it finds the {{}}
			if result, err := worker(ctx, db, sdb, accountID, work.Expression, variables); err != nil {
				return err.Error()
			} else {
				work.InboundRespCh <- result
			}
		case ruler.Querier: //input:executes this when it finds the <<>>
			if result, err := querier(ctx, db, accountID, work.Expression, variables); err != nil {
				return err.Error()
			} else {
				work.InboundRespCh <- result
			}
		case ruler.Computer: //output:executes this when it finds EOF/Response
			lexedContent = work.OutboundResp
		}
	}
	return lexedContent
}

// RunExpGrapher run the expression and returns graph query in a readable format
func (e *Engine) RunExpGrapher(ctx context.Context, db *sqlx.DB, sdb *database.SecDB, accountID, exp string) *ruler.Filter {
	var filter *ruler.Filter
	signalsChan := make(chan ruler.Work)
	go ruler.Run(exp, ruler.Graph, signalsChan)
	//signalsChan wait to receive evaluation work and final evaluated string
	for work := range signalsChan {
		switch work.Type {
		case ruler.Worker: //why worker calling grapher? because the logic is same as worker
			if result, err := grapher(ctx, db, sdb, accountID, work.Expression); err != nil {
				log.Printf("***> unexpected error occurred. sending empty conditions - error: %v ", err)
				return nil
			} else {
				work.InboundRespCh <- result
			}
		case ruler.Grapher: //output:executes this when it finds EOF/Response
			filter = work.OutboundResp.(*ruler.Filter)
		}
	}
	return filter
}

// RunExpEvaluator runs the expression to see whether the condition met or not
func (e *Engine) RunExpEvaluator(ctx context.Context, db *sqlx.DB, sdb *database.SecDB, accountID, exp string, variables map[string]interface{}) bool {
	positive := false
	signalsChan := make(chan ruler.Work)
	go ruler.Run(exp, ruler.Execute, signalsChan)
	//signalsChan wait to receive evaluation work and final execution
	for work := range signalsChan {
		switch work.Type {
		case ruler.Worker: //input:executes this when it finds the {{}}
			if result, err := worker(ctx, db, sdb, accountID, work.Expression, variables); err != nil {
				log.Println("***> unexpected error occurred when evaluting the expression. error: ", err)
				return false
			} else {
				work.InboundRespCh <- result
			}
		case ruler.Grapher: //input:executes this when it finds the what??? - actually here the grapher should come into picture when the conditions refers the reference value. Need to implement
			if result, err := grapher(ctx, db, sdb, accountID, work.Expression); err != nil {
				log.Println("***> unexpected error occurred when evaluting the expression. error: ", err)
				return false
			} else {
				work.InboundRespCh <- result
			}
		case ruler.PosExecutor: //output:executes this when it finds EOF/Response
			positive = true
		}
	}
	log.Printf("internal.rule.engine: result of the evaluator: %t\n", positive)
	return positive
}

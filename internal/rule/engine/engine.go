package engine

import (
	"context"
	"log"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/jobber"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/ruler"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
)

type Engine struct {
	Job jobber.Jobber
}

// RuleResult returns back the recently executed results
type RuleResult struct {
	Executed bool
	Response map[string]interface{}
}

//RunRuleEngine runs the expression and execute action if the expression conditions met
func (e *Engine) RunRuleEngine(ctx context.Context, db *sqlx.DB, rp *redis.Pool, n node.Node) (*RuleResult, error) {
	var err error
	signalsChan := make(chan ruler.Work)
	go ruler.Run(n.Expression, ruler.Execute, signalsChan)
	ruleResult := &RuleResult{}
	//signalsChan wait to receive evaluation work and final execution
	for work := range signalsChan {
		switch work.Type {
		case ruler.Worker:
			if result, err := worker(ctx, db, n.AccountID, work.Expression, n.VariablesMap()); err != nil {
				return nil, err
			} else {
				work.InboundRespCh <- result
			}
		case ruler.Querier:
			if result, err := querier(ctx, db, rp, n.AccountID, work.Expression); err != nil {
				return nil, err
			} else {
				work.InboundRespCh <- result
			}
		case ruler.PosExecutor:
			err = ruleResult.executePosCase(ctx, db, n, e)
		case ruler.NegExecutor:
			err = ruleResult.executeNegCase(ctx, db, n)
		}
	}

	return ruleResult, err
}

//RunExpRenderer run the expression and returns evaluated string
func (e *Engine) RunExpRenderer(ctx context.Context, db *sqlx.DB, accountID, exp string, variables map[string]interface{}) string {
	var lexedContent string
	signalsChan := make(chan ruler.Work)
	go ruler.Run(exp, ruler.Parse, signalsChan)
	//signalsChan wait to receive evaluation work and final evaluated string
	for work := range signalsChan {
		switch work.Type {
		case ruler.Worker:
			if result, err := worker(ctx, db, accountID, work.Expression, variables); err != nil {
				return err.Error()
			} else {
				work.InboundRespCh <- result
			}
		case ruler.Parser:
			lexedContent = work.OutboundResp.(string)
		}
	}
	return lexedContent
}

//RunExpQuerier run the expression and returns conditions in a readable format
func (e *Engine) RunExpQuerier(ctx context.Context, db *sqlx.DB, rp *redis.Pool, accountID, exp string) []ruler.Condition {
	var conditions []ruler.Condition
	signalsChan := make(chan ruler.Work)
	go ruler.Run(exp, ruler.Query, signalsChan)
	//signalsChan wait to receive evaluation work and final evaluated string
	for work := range signalsChan {
		switch work.Type {
		case ruler.Worker: //why worker calling querier? because the logic is same as worker
			if result, err := querier(ctx, db, rp, accountID, work.Expression); err != nil {
				log.Println("err occurred. Sending empty conditions - ", err)
				return []ruler.Condition{}
			} else {
				work.InboundRespCh <- result
			}
		case ruler.Querier:
			conditions = work.OutboundResp.([]ruler.Condition)
		}
	}
	return conditions
}

//RunExpEvaluator runs the expression to see whether the condition met or not
func (e *Engine) RunExpEvaluator(ctx context.Context, db *sqlx.DB, rp *redis.Pool, accountID, exp string, variables map[string]interface{}) bool {
	positive := false
	signalsChan := make(chan ruler.Work)
	go ruler.Run(exp, ruler.Execute, signalsChan)
	//signalsChan wait to receive evaluation work and final execution
	for work := range signalsChan {
		switch work.Type {
		case ruler.Worker:
			if result, err := worker(ctx, db, accountID, work.Expression, variables); err != nil {
				log.Println("error in expression evaluator ", err)
				return false
			} else {
				work.InboundRespCh <- result
			}
		case ruler.Querier:
			if result, err := querier(ctx, db, rp, accountID, work.Expression); err != nil {
				log.Println("error in expression evaluator ", err)
				return false
			} else {
				work.InboundRespCh <- result
			}
		case ruler.PosExecutor:
			positive = true
		}
	}
	log.Printf("result of the evaluator: %t", positive)
	return positive
}

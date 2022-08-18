package engine

import (
	"context"
	"strings"

	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"gitlab.com/vjsideprojects/relay/internal/user"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
)

func grapher(ctx context.Context, db *sqlx.DB, rp *redis.Pool, accountID, expression string) (interface{}, error) {
	//log.Println("rule.engine.grapher:  query: ", expression)
	if expression == node.MeEntity { //just like making grapher smart
		currentUserID, err := user.RetrieveCurrentUserID(ctx)
		if err != nil {
			return nil, err
		}
		return memberID(ctx, db, accountID, currentUserID)
	}
	elements := strings.Split(expression, ".")
	return elements[1], nil
}

//Not used for now. Could be useful in the future. If we decided to execute rules of inside worker
func gSegmentJson(rp *redis.Pool, expression string, input map[string]interface{}) (interface{}, error) {
	gSegment, err := graphdb.GraphNodeSt(expression)
	if err != nil {
		return nil, err
	}
	if itemID, ok := input[gSegment.Label]; ok {
		gSegment = gSegment.AddIDCondition(itemID)
		qr, err := graphdb.GetResult(rp, gSegment, 0, "", "")

		if err != nil {
			return nil, err
		}
		if !qr.Empty() {
			return true, nil
		}
	}
	return nil, nil
}

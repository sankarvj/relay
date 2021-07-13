package engine

import (
	"context"
	"log"
	"strings"

	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
)

func querier(ctx context.Context, db *sqlx.DB, rp *redis.Pool, accountID, expression string) (interface{}, error) {
	log.Println("query ---> ", expression)
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
		qr, err := graphdb.GetResult(rp, gSegment)

		if err != nil {
			return nil, err
		}
		if !qr.Empty() {
			return true, nil
		}
	}
	return nil, nil
}

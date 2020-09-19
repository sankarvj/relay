package engine

import (
	"context"
	"log"

	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"

	"github.com/gomodule/redigo/redis"
)

func querier(ctx context.Context, rp *redis.Pool, query string, input map[string]interface{}) (map[string]interface{}, error) {
	log.Println("query ---> ", query)
	gSegment, err := graphdb.GraphNodeSt(query)
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
			return map[string]interface{}{"result": true}, nil
		}
	}

	return map[string]interface{}{}, nil
}

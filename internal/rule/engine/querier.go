package engine

import (
	"context"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
)

//querier is used to render the actual values from the template field values
//ex: 1 days from creation to actual date
//in future we can use querier for further evaluation
func querier(ctx context.Context, db *sqlx.DB, accountID string, expression string, input map[string]interface{}) (interface{}, error) {
	log.Printf("internal.rule.engine.querier running expression: %s\n", expression)

	x := util.ConvertStrToInt(expression)
	t := time.Now()
	addedDate := t.AddDate(0, 0, x)
	//return time here
	return util.FormatTimeGo(addedDate), nil
}

package engine

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func executeHook(ctx context.Context, db *sqlx.DB, entityFields []entity.Field) (map[string]interface{}, error) {
	result, err := retriveAPIEntityResult(entityFields)
	log.Println("result :: ", result)
	log.Println("err :: ", err)
	return result, err
}
